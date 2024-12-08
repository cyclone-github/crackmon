package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

/*
v2023-10-07.1520
v2023-10-08.0930; fixed https://github.com/cyclone-github/crackmon/issues/3
v2023-10-13.1445; fixed https://github.com/cyclone-github/crackmon/issues/4; refactored sendX commands
*/

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

// watch for ctrl+c
func catchCtrlC(stdin io.Writer) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT) // SIGINT = Ctrl+C

	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "\nCaught Ctrl+C. Shutting down...\n")
		// sendQ(stdin)
		// time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}

var (
	sendB func(io.Writer)
	sendQ func(io.Writer)
	wg    sync.WaitGroup
	done  = make(chan struct{})
)

func initializeAndExecuteCommon(cmdStr string, timeT int, crackT int, debug bool, stdin io.Writer, stdout io.Reader, osChecker func() string) {
	catchCtrlC(stdin)

	scanner := bufio.NewScanner(stdout)

	var (
		hashcatRunning   bool
		hashcatPaused    bool
		hashcatStartTime time.Time
		missedChecks     int
		hashcatStatus    string
		stdoutStatus     string
		cumulativeAvg    float64
		totalCracks      int
		totalTime        float64
	)
	// regex for reading hc stdout
	runningRe := regexp.MustCompile(`Status\.+: Running`)
	pausedRe := regexp.MustCompile(`Status\.+: Paused`)
	stoppedRe := regexp.MustCompile(`Stopped:`)
	recoveredRe := regexp.MustCompile(`Recovered[.]+:\s*(\d+)`)
	dictCacheRe := regexp.MustCompile(`Dictionary cache building`)
	invalidArgRe := regexp.MustCompile(`Invalid argument specified.`)

	// default hc status
	hashcatStatus = "Waiting for status"
	stdoutStatus = "Waiting for stdout"
	hashcatRunning = false
	hashcatPaused = false

	wg.Add(2)

	go func() {
		time.Sleep(10 * time.Millisecond)
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				if scanner.Scan() {
					line := scanner.Text()
					fmt.Println(line)

					// detect "hashcat : unknown option"
					if invalidArgRe.MatchString(line) {
						if debug {
							fmt.Fprintln(os.Stderr, "Invalid argument specified. Exiting...")
						}
						os.Exit(1)
						return
						// check if hc is running
					} else if runningRe.MatchString(line) {
						hashcatStatus = "Running"
						hashcatRunning = true
						hashcatPaused = false
						missedChecks = 0
						if hashcatStartTime.IsZero() { // set hc start time
							hashcatStartTime = time.Now().Add(-time.Second * 9)
						}
						// check is hc is paused
					} else if pausedRe.MatchString(line) {
						hashcatStatus = "Paused"
						hashcatPaused = true
						hashcatRunning = false
						missedChecks = 0
						// check is hc is stopped
					} else if stoppedRe.MatchString(line) {
						close(done)
						time.Sleep(1 * time.Second)
						os.Exit(0)
						return
						// Check for "Dictionary cache building"
						// not working due to hashcat stdout uses \r rather than \n during dictionary cache building
					} else if !hashcatRunning && dictCacheRe.MatchString(line) {
						hashcatStatus = "Building dictionary cache..."
						hashcatRunning = true
						hashcatPaused = false
						stdoutStatus = "OK"
						missedChecks = 0
					}
					// check founds total
					recoveredT := recoveredRe.FindStringSubmatch(line)
					if len(recoveredT) >= 2 && hashcatRunning && !hashcatPaused {
						totalCracks, _ = strconv.Atoi(recoveredT[1])
						totalTime = time.Since(hashcatStartTime).Seconds()
						stdoutStatus = "OK"

						// cumulative average, total cracks / total time (same as hashcat's AVG)
						if totalTime > 60 {
							cumulativeAvg = float64(totalCracks) / totalTime * 60
						}

						// sendB if -c threshold is not met within -t threshold
						if totalTime >= float64(timeT*60) {
							if cumulativeAvg < float64(crackT) {
								sendB(stdin)
							}
						}
					}
				} else {
					// sendQ if hc stdout cannot be read
					missedChecks++
					hashcatStatus = "Unknown"
					stdoutStatus = "Cannot read stdout"
					if missedChecks >= 120 {
						sendQ(stdin)
						time.Sleep(1 * time.Second)
						os.Exit(1)
						return
					}
				}
			}
		}
	}()

	// debug output goroutine
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10500 * time.Millisecond)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if debug {
					fmt.Fprintf(os.Stderr, "DEBUG: hashcat status:\t %s\n", hashcatStatus)
					fmt.Fprintf(os.Stderr, "DEBUG: hashcat stdout:\t %s\n", stdoutStatus)
					fmt.Fprintf(os.Stderr, "DEBUG: Recovered:\t %d\n", totalCracks)
					if totalTime < 60 {
						fmt.Fprintf(os.Stderr, "DEBUG: Recovered AVG:\t N/A\n\n")
					} else {
						fmt.Fprintf(os.Stderr, "DEBUG: Recovered AVG:\t %.1f\n\n", cumulativeAvg)
					}
				}
			}
		}
	}()
	wg.Wait()
	close(done)
}
