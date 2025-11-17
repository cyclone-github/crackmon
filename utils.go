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
func catchCtrlC(io.Writer) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT) // SIGINT = Ctrl+C

	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "\nCaught Ctrl+C. Shutting down...")
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

func initializeAndExecuteCommon(timeT int, crackT int, debug bool, stdin io.Writer, stdout io.Reader) {
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
		initialRecovered int
		baselineSet      bool

		// mdxfind-specific state
		mdxStartTime time.Time
	)

	// mutex to protect debug-visible fields
	var statsMu sync.Mutex

	// regex for reading hashcat stdout
	runningRe := regexp.MustCompile(`Status\.+: Running`)
	pausedRe := regexp.MustCompile(`Status\.+: Paused`)
	stoppedRe := regexp.MustCompile(`Stopped:`)
	recoveredRe := regexp.MustCompile(`Recovered[.]+:\s*(\d+)`)
	removedRe := regexp.MustCompile(`INFO: Removed (\d+) hashes`)
	dictCacheRe := regexp.MustCompile(`Dictionary cache building`)
	invalidArgRe := regexp.MustCompile(`Invalid argument specified.`)

	// regex for reading mdxfind stdout
	mdxWorkingRe := regexp.MustCompile(`^Working on `)
	mdxFoundRe := regexp.MustCompile(`Found=(\d+)`)
	mdxDoneRe := regexp.MustCompile(`^Done - `)
	mdxTotalFoundRe := regexp.MustCompile(`Total hashes found`)

	// default status
	statsMu.Lock()
	hashcatStatus = "Waiting for status"
	stdoutStatus = "Waiting for stdout"
	statsMu.Unlock()
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

					switch currentRunner {
					case RunnerHashcat:
						// detect "hashcat : unknown option"
						if invalidArgRe.MatchString(line) {
							if debug {
								fmt.Fprintln(os.Stderr, "Invalid argument specified. Exiting...")
							}
							os.Exit(1)
							return
							// check if hc is running
						} else if runningRe.MatchString(line) {
							statsMu.Lock()
							hashcatStatus = "Running"
							statsMu.Unlock()
							hashcatRunning = true
							hashcatPaused = false
							missedChecks = 0
							if hashcatStartTime.IsZero() { // set hc start time
								hashcatStartTime = time.Now().Add(-time.Second * 9)
							}
							// check is hc is paused
						} else if pausedRe.MatchString(line) {
							statsMu.Lock()
							hashcatStatus = "Paused"
							statsMu.Unlock()
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
							statsMu.Lock()
							hashcatStatus = "Building dictionary cache..."
							stdoutStatus = "OK"
							statsMu.Unlock()
							hashcatRunning = true
							hashcatPaused = false
							missedChecks = 0
						}

						// set baseline zero from "INFO: Removed" line
						removedT := removedRe.FindStringSubmatch(line)
						if len(removedT) >= 2 && !baselineSet {
							initialRecovered, _ = strconv.Atoi(removedT[1])
							baselineSet = true
							if debug {
								fmt.Fprintf(os.Stderr, "DEBUG: Baseline zero set from INFO: Removed: %d\n", initialRecovered)
							}
						}
						// check founds total
						recoveredT := recoveredRe.FindStringSubmatch(line)
						if len(recoveredT) >= 2 && hashcatRunning && !hashcatPaused {
							currentRecovered, _ := strconv.Atoi(recoveredT[1])

							actualRecovered := currentRecovered
							if baselineSet {
								actualRecovered -= initialRecovered
							}

							totalTime = time.Since(hashcatStartTime).Seconds()

							// cumulative average, total cracks / total time (same as hashcat's AVG)
							if totalTime > 60 {
								cumulativeAvg = float64(actualRecovered) / totalTime * 60
							}

							// recovered minus existing founds in potfile
							totalCracks = actualRecovered

							// update debug-visible fields
							statsMu.Lock()
							stdoutStatus = "OK"
							if totalTime > 60 {
							}
							statsMu.Unlock()

							// sendB if -c threshold is not met within -t threshold
							if totalTime >= float64(timeT*60) {
								if cumulativeAvg < float64(crackT) {
									sendB(stdin)
								}
							}
						}

					case RunnerMDXFind:
						// treat "Working on ..." as running status
						if mdxWorkingRe.MatchString(line) {
							statsMu.Lock()
							hashcatStatus = "Running"
							stdoutStatus = "OK"
							statsMu.Unlock()
							hashcatRunning = true
							hashcatPaused = false
							missedChecks = 0
							if mdxStartTime.IsZero() {
								mdxStartTime = time.Now()
							}
						}

						// parse Found=N to track cumulative cracks
						foundT := mdxFoundRe.FindStringSubmatch(line)
						if len(foundT) >= 2 {
							currentFound, _ := strconv.Atoi(foundT[1])

							actualFound := currentFound
							if actualFound < 0 {
								actualFound = 0
							}

							if mdxStartTime.IsZero() {
								mdxStartTime = time.Now()
							}
							totalTime = time.Since(mdxStartTime).Seconds()

							if totalTime > 0 {
								cumulativeAvg = float64(actualFound) / totalTime * 60
							}

							totalCracks = actualFound

							// update debug-visible fields
							statsMu.Lock()
							stdoutStatus = "OK"
							statsMu.Unlock()

							if totalTime >= float64(timeT*60) {
								if cumulativeAvg < float64(crackT) {
									sendB(stdin) // for mdxfind: Ctrl+C
								}
							}

						}

						// detect end of mdxfind run
						if mdxDoneRe.MatchString(line) || mdxTotalFoundRe.MatchString(line) {
							close(done)
							time.Sleep(1 * time.Second)
							os.Exit(0)
							return
						}
					default:
						// unknown Runner; do nothing special
					}

				} else {
					// sendQ if stdout cannot be read
					missedChecks++
					statsMu.Lock()
					hashcatStatus = "Unknown"
					stdoutStatus = "Cannot read stdout"
					statsMu.Unlock()
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
					statsMu.Lock()
					status := hashcatStatus
					out := stdoutStatus
					cracks := totalCracks
					avg := cumulativeAvg
					statsMu.Unlock()

					fmt.Fprintf(os.Stderr, "DEBUG: status:\t\t %s\n", status)
					fmt.Fprintf(os.Stderr, "DEBUG: stdout:\t\t %s\n", out)
					fmt.Fprintf(os.Stderr, "DEBUG: Recovered:\t %d\n", cracks)
					fmt.Fprintf(os.Stderr, "DEBUG: Recovered AVG:\t %.1f\n\n", avg)
				}
			}
		}
	}()
	wg.Wait()
	close(done)
}
