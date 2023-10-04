//go:build linux || darwin
// +build linux darwin

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// coded by cyclone
// tested on linux, should also work on mac
// v2023-10-04.1545

var (
	sendB func(io.Writer)
	sendQ func(io.Writer)
)

func linuxSendB(stdin io.Writer) {
	io.WriteString(stdin, "b\n")
}

func linuxSendQ(stdin io.Writer) {
	io.WriteString(stdin, "q\n")
}

func initializeAndExecute(cmdStr string, timeT int, crackT int, re *regexp.Regexp, debug bool) {
	cmdSlice := strings.Fields(cmdStr)
	cmdName := cmdSlice[0]
	cmdArgs := cmdSlice[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	stdin, _ := cmd.StdinPipe()
	cmd.Stderr = os.Stderr
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting command:", err)
		return
	}

	sendB = linuxSendB
	sendQ = linuxSendQ

	scanner := bufio.NewScanner(stdout)
	var recoveredTimeSeen bool = false

	// monitor hashcat output
	go func() {
		lastCrackTime := time.Now()
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				cracks, _ := strconv.Atoi(matches[1])
				if cracks < crackT {
					elapsed := time.Since(lastCrackTime).Minutes()
					if int(elapsed) >= timeT {
						if debug {
							fmt.Fprintf(os.Stderr, "\nDEBUG: cracks = %d, elapsed = %f, timeT = %d\n", cracks, elapsed, timeT)
							fmt.Fprintln(os.Stderr, "DEBUG: Hashcat bypass ('b') sent\n")
						}
						sendB(stdin)
						lastCrackTime = time.Now()
					}
				} else {
					lastCrackTime = time.Now()
				}
			}

			if strings.Contains(line, "Recovered/Time") {
				recoveredTimeSeen = true
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "stdin error:", err)
		}
	}()

	// debugging goroutine
	go func() {
		time.Sleep(10 * time.Second) // give hashcat time to start
		ticker := time.NewTicker(1 * time.Second)
		missedChecks := 0
		for range ticker.C {
			if !recoveredTimeSeen {
				missedChecks++
				if debug {
					fmt.Fprintf(os.Stderr, "DEBUG: Cannot read hashcat output...\r")
				}
				if missedChecks >= 120 { // stop if can't read hashcat for 120 checks / seconds
					if debug {
						fmt.Fprintf(os.Stderr, "DEBUG: Stopping, unable to read hashcat output for %d seconds.\n", missedChecks)
						fmt.Fprintln(os.Stderr, "DEBUG: Program must be able to read hashcat output: 'Recovered/Time...: CUR:'\n")
					}
					sendQ(stdin)
					os.Exit(0)
				}
			} else {
				if debug {
					fmt.Fprint(os.Stderr, "DEBUG: Hashcat output: OK\r")
				}
				missedChecks = 0
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// wait for hashcat to finish
	cmd.Wait()
	if debug {
		fmt.Fprintln(os.Stderr, "\nDEBUG: Hashcat process has stopped. Exiting...\n")
	}
}
