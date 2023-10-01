package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

/*
crackmon - by cyclone
original idea by: https://github.com/justpretending/avgdrop
hashcat wrapper tool to similate pressing "b" key to bypass current hashcat attack if cracking rate goes below threshold
developed and tested on debian 12 linux
should work fine on mac OS X and later, but not tested
does not curently work on windows due to windows having crappy PTY support (cannot send keyboard commands to shell)
currently uses "CUR" rather than "AVG" crack rate (can be changed by modifying regex "re := regexp.MustCompile(`CUR:(\d+),`)" on line 127)

example:
./crackmon ./hashcat {hashcat args} 			(defaults to -time 1m -crack 1)
./crackmon -t 10 -c 2 ./hashcat {hashcat args} 	(custom -time 10m -crack 2)

version:
v2023-10-01.1700; initial github release
*/

func help() {
	version()
	fmt.Fprintln(os.Stderr, "\nExamples:\n")
	fmt.Fprintln(os.Stderr, "(Defaults to -time 1m -crack 1)")
	fmt.Fprintln(os.Stderr, "./crackmon ./hashcat {hashcat args}\n")
	fmt.Fprintln(os.Stderr, "Custom: -time 10m -crack 2")
	fmt.Fprintln(os.Stderr, "./crackmon -t 10 -c 2 ./hashcat {hashcat args}\n")
}

func version() {
	fmt.Fprintln(os.Stderr, "crackmon v2023-10-01.1700; initial github release")
	fmt.Fprintln(os.Stderr, "https://github.com/cyclone-github/crackmon")
	fmt.Fprintln(os.Stderr, "Original idea by: https://github.com/justpretending/avgdrop")
	detectedOS := checkOS()
	fmt.Fprintf(os.Stderr, "\nDetected OS: %s\n", detectedOS)
}

func cyclone() {
	codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
	codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
	fmt.Fprintln(os.Stderr, string(codedByDecoded))
}

func checkOS() string {
	return runtime.GOOS
}

func main() {
	timeThreshold := flag.Int("t", 1, "Time threshold in minutes")
	cracksThreshold := flag.Int("c", 1, "Cracks per time threshold")
	cycloneFlag := flag.Bool("cyclone", false, "Display message")
	versionFlag := flag.Bool("version", false, "Display version")
	helpFlag := flag.Bool("help", false, "Display help")
	flag.Parse()
	detectedOS := checkOS()

	if *cycloneFlag {
		cyclone()
		os.Exit(0)
	}

	if *versionFlag {
		version()
		os.Exit(0)
	}

	if *helpFlag {
		help()
		os.Exit(0)
	}

	// capture remaining hashcat arguments to execute
	cmdArgs := flag.Args()

	// check if hashcat arguments are missing
	if len(cmdArgs) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing hashcat command to execute.\n")
		help()
		return
	}

	// add --status-timer=10, --status, and --advice-disable if not present
	cmdStr := strings.Join(cmdArgs, " ")

	if !strings.Contains(cmdStr, "--status") {
		cmdStr += " --status"
	}
	if !strings.Contains(cmdStr, "--status-timer") {
		cmdStr += " --status-timer=10"
	}
	if !strings.Contains(cmdStr, "--advice-disable") {
		cmdStr += " --advice-disable"
	}

	// print the full command to be executed
	fmt.Fprintln(os.Stderr, "\nExecuting:", cmdStr+"\n")

	// split the full command string into command and arguments
	cmdSlice := strings.Fields(cmdStr)
	cmdName := cmdSlice[0]
	cmdArgs = cmdSlice[1:]

	// create command execution
	cmd := exec.Command(cmdName, cmdArgs...)
	stdin, _ := cmd.StdinPipe()
	cmd.Stderr = os.Stderr
	stdout, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting command:", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	re := regexp.MustCompile(`CUR:(\d+),`)
	counter := 0

	// "Recovered/Time" var
	var recoveredTimeSeen bool = false

	// determine appropriate line ending per OS (Windows is not currently working)
	var lineEnding string
	if detectedOS == "windows" {
		lineEnding = "\r\n"
	} else {
		lineEnding = "\n"
	}

	// monitor hashcat output
	go func() {
		fmt.Fprintf(os.Stderr, "Debug: Starting on OS: %s\n\n", detectedOS)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // print full hashcat command -- debugging
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				cracks, _ := strconv.Atoi(matches[1])
				if cracks < *cracksThreshold {
					counter++
					if counter >= *timeThreshold {
						fmt.Fprintln(os.Stderr, "\nDebug: Hashcat bypass sent\n")
						stdin.Write([]byte("b" + lineEnding))
						counter = 0
					}
				} else {
					counter = 0
				}
			}

			if strings.Contains(line, "Recovered/Time") {
				recoveredTimeSeen = true
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "stdin error:", err) // debug
		}
	}()

	// debugging goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		missedChecks := 0 // counter for missed checks
		for range ticker.C {
			if !recoveredTimeSeen {
				fmt.Fprintln(os.Stderr, "Debug: Hashcat output: FALSE\n")
				missedChecks++
				if missedChecks >= 3 {
					fmt.Fprintln(os.Stderr, "Debug: Stopping Hashcat due to 3 consecutive missed checks.\n")
					stdin.Write([]byte("q" + lineEnding))
				}
			} else {
				fmt.Fprintln(os.Stderr, "Debug: Hashcat output: OK\n")
				missedChecks = 0
			}
		}
	}()

	// wait for command to finish
	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "hashcat command error:", err)
	}
}

// end code
