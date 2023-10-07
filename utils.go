package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// v2023-10-07.1520

func ReadUserInput(stdin io.Writer) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		io.WriteString(stdin, text)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
	}
}

func checkOS() string {
	return runtime.GOOS
}

var (
	sendB func(io.Writer)
	sendQ func(io.Writer)
	wg    sync.WaitGroup
	done  = make(chan struct{})
)

func initializeAndExecuteCommon(cmdStr string, timeT int, crackT int, re *regexp.Regexp, debug bool, stdin io.Writer, stdout io.Reader, osChecker func() string) {
	scanner := bufio.NewScanner(stdout)
	var recoveredTimeSeen bool
	var hashcatRunning bool
	var hashcatPaused bool

	runningRe := regexp.MustCompile(`Status\.+: Running`)
	pausedRe := regexp.MustCompile(`Status\.+: Paused`)
	stoppedRe := regexp.MustCompile(`Stopped:`)

	wg.Add(2)

	// check hashcat output and sending commands
	go func() {
		defer wg.Done()
		lastCrackTime := time.Now()
		for {
			select {
			case <-done:
				return
			default:
				if scanner.Scan() {
					line := scanner.Text()
					fmt.Println(line)

					// check if Hashcat is running or paused
					if runningRe.MatchString(line) {
						hashcatRunning = true
						hashcatPaused = false
					} else if pausedRe.MatchString(line) {
						hashcatPaused = true
						hashcatRunning = false
					}

					// check if Hashcat is stopped
					if stoppedRe.MatchString(line) {
						if debug {
							fmt.Fprintf(os.Stderr, "DEBUG: Hashcat has stopped. Exiting...\n")
						}
						close(done)
						time.Sleep(1 * time.Second)
						os.Exit(0)
						return
					}

					// check for AVG cracks
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						cracks, _ := strconv.Atoi(matches[1])
						if cracks < crackT && hashcatRunning && !hashcatPaused {
							elapsed := time.Since(lastCrackTime).Minutes()
							if int(elapsed) >= timeT {
								if debug {
									fmt.Fprintf(os.Stderr, "\nDEBUG: cracks = %d, elapsed = %f, timeT = %d\n", cracks, elapsed, timeT)
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
				}
			}
		}
	}()

	// goroutine for debug output
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		missedChecks := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if hashcatRunning {
					if !recoveredTimeSeen && hashcatRunning && !hashcatPaused {
						missedChecks++
						if debug {
							fmt.Fprintf(os.Stderr, "DEBUG: Cannot read hashcat output...\r")
						}
						if missedChecks >= 60 {
							if debug {
								fmt.Fprintf(os.Stderr, "DEBUG: Stopping, unable to read hashcat output for %d seconds.\n", missedChecks)
							}
							sendQ(stdin)
							close(done)
							return
						}
					} else {
						if debug {
							fmt.Fprint(os.Stderr, "DEBUG: Hashcat output: OK\r")
						}
						missedChecks = 0
					}
				}
			}
		}
	}()

	// wait for all goroutines to complete
	wg.Wait()
	close(done)
}
