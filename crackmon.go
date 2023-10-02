package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
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
v2023-10-02.1030-debug; fixed timeThreshold; added more debugging info
*/

func help() {
	version()
	fmt.Fprintln(os.Stderr, "\nExamples:\n")
	fmt.Fprintln(os.Stderr, "(Defaults to -time 5m -crack 1)")
	fmt.Fprintln(os.Stderr, "./crackmon ./hashcat {hashcat args}\n")
	fmt.Fprintln(os.Stderr, "Custom: -time 10m -crack 2")
	fmt.Fprintln(os.Stderr, "./crackmon -t 10 -c 2 ./hashcat {hashcat args}\n")
}

func version() {
	fmt.Fprintln(os.Stderr, "crackmon v2023-10-02.1030-debug")
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
	timeThreshold := flag.Int("t", 5, "Time threshold in minutes")
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
	fmt.Fprintf(os.Stderr, "DEBUG: timeThreshold = %d\n", *timeThreshold)
	fmt.Fprintf(os.Stderr, "DEBUG: cracksThreshold = %d\n", *cracksThreshold)
	time.Sleep(1 * time.Second)

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
	//counter := 0

	// "Recovered/Time" var
	var recoveredTimeSeen bool = false

	// determine appropriate line ending per OS (Windows is not currently working)
	var lineEnding string
	if detectedOS == "windows" {
		lineEnding = "\r\n"
		//lineEnding = "\n"
	} else {
		lineEnding = "\n"
	}

	// monitor hashcat output
	go func() {
		fmt.Fprintf(os.Stderr, "DEBUG: Starting on OS: %s\n\n", detectedOS)
		lastCrackTime := time.Now() // start time
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // print full hashcat command -- debugging
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				cracks, _ := strconv.Atoi(matches[1])
				if cracks < *cracksThreshold {
					elapsed := time.Since(lastCrackTime).Minutes() // get elapsed time in minutes
					if int(elapsed) >= *timeThreshold {
						fmt.Fprintf(os.Stderr, "\nDEBUG: cracks = %d, elapsed = %f, timeThreshold = %d\n", cracks, elapsed, *timeThreshold)
						fmt.Fprintln(os.Stderr, "DEBUG: Hashcat bypass ('b') sent\n")
						io.WriteString(stdin, "b"+lineEnding) // io.WriteString works on linux and windows
						time.Sleep(1 * time.Second)           // add delay (testing: for windows)
						lastCrackTime = time.Now()            // reset crack time
					}
				} else {
					lastCrackTime = time.Now() // reset crack time
				}
			}

			if strings.Contains(line, "Recovered/Time") {
				recoveredTimeSeen = true
			}
			time.Sleep(10 * time.Millisecond) // add delay
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "stdin error:", err) // debug
		}
	}()

	// debugging goroutine
	go func() {
		time.Sleep(10 * time.Second) // add delay to allow hashcat to start
		ticker := time.NewTicker(10 * time.Second)
		missedChecks := 0 // counter for missed checks
		for range ticker.C {
			fmt.Fprintln(os.Stderr, "DEBUG: Ticker: OK")

			if !recoveredTimeSeen {
				missedChecks++
				fmt.Fprintf(os.Stderr, "DEBUG: missedChecks: %d\n", missedChecks)
				fmt.Fprintln(os.Stderr, "DEBUG: Hashcat output: FALSE\n")

				if missedChecks >= 3 {
					fmt.Fprintln(os.Stderr, "DEBUG: Stopping Hashcat due to 3 consecutive missed checks.")
					fmt.Fprintln(os.Stderr, "DEBUG: Program must be able to read hashcat output: 'Recovered/Time...: CUR:'\n")
					io.WriteString(stdin, "q"+lineEnding) // io.WriteString works on linux and windows
					time.Sleep(1 * time.Second)           // add delay (testing: for windows)
				}
			} else {
				fmt.Fprintln(os.Stderr, "DEBUG: Hashcat output: OK\n")
				missedChecks = 0
			}
			time.Sleep(100 * time.Millisecond) // add delay
		}
	}()

	// wait for command to finish
	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "hashcat status:", err)
	}
}

// end code
