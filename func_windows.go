//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UserExistsError/conpty"
)

// coded by cyclone
// uses a windows api wrapper for terminal PTY support
// not working yet
// v2023-10-04.1545

var (
	cptyInstance *conpty.ConPty
	sendB        func(io.Writer)
	sendQ        func(io.Writer)
)

func initializeConPTY(fullCmd string) error {
	var err error
	cptyInstance, err = conpty.Start(fullCmd, conpty.ConPtyDimensions(80, 40))
	if err != nil {
		return err
	}
	// fmt.Fprintf(os.Stderr, "DEBUG: Running: %s\n", fullCmd)
	return nil
}

func windowsSendB(stdin io.Writer) {
	if cptyInstance == nil {
		fmt.Fprintln(os.Stderr, "ConPTY instance not initialized.")
		return
	}
	_, err := cptyInstance.Write([]byte("b\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send 'b' command: %v", err)
	}
}

func windowsSendQ(stdin io.Writer) {
	if cptyInstance == nil {
		fmt.Fprintf(os.Stderr, "ConPTY instance not initialized.")
		return
	}
	_, err := cptyInstance.Write([]byte("q\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send 'q' command: %v", err)
	} else {
		cptyInstance.Close()
		//close(done) // signal all goroutines to close
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}
}

func initializeAndExecute(cmdStr string, timeT int, crackT int, re *regexp.Regexp, debug bool) {
	var wg sync.WaitGroup
	done := make(chan struct{})
	var recoveredTimeSeen bool = false

	err := initializeConPTY(cmdStr) // full hashcat command
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing ConPTY:", err)
		return
	}
	sendB = windowsSendB
	sendQ = windowsSendQ

	wg.Add(3)

	// monitor hashcat output
	go func() {
		defer wg.Done()
		lastCrackTime := time.Now()
		buf := make([]byte, 1024)

		for {
			select {
			case <-done:
				return
			default:
				n, err := cptyInstance.Read(buf)
				if err != nil {
					// fmt.Fprintln(os.Stderr, "Error reading output:", err)
					return
				}
				if n > 0 {
					line := string(buf[:n])
					fmt.Fprint(os.Stderr, line)
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
								sendB(cptyInstance)
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
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// debug goroutine
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Second) // give hashcat time to start
		ticker := time.NewTicker(1 * time.Second)
		missedChecks := 0
		for range ticker.C {
			select {
			case <-done:
				return
			default:
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
						sendQ(cptyInstance)
					}
				} else {
					if debug {
						fmt.Fprint(os.Stderr, "DEBUG: Hashcat output: OK\r")
					}
					missedChecks = 0
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	// monitor if hashcat is still running
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				time.Sleep(100 * time.Millisecond)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				exitCode, err := cptyInstance.Wait(ctx)
				cancel()

				if err == nil && exitCode != 259 { // 259 is STILL_ACTIVE
					if debug {
						fmt.Fprintln(os.Stderr, "\nDEBUG: Hashcat process has stopped. Exiting...\n")
					}
					cptyInstance.Close()
					close(done) // signal all goroutines to close
					//break
				}
			}
		}
	}()
	wg.Wait() // wait for all goroutines to complete
}
