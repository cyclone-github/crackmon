package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

/*
GNU GENERAL PUBLIC LICENSE Version 2, June 1991
https://github.com/cyclone-github/crackmon/blob/main/LICENSE

crackmon - by cyclone
original idea by: https://github.com/justpretending/avgdrop
hashcat wrapper tool to similate pressing "b" key to bypass current hashcat attack if cracking rate goes below threshold
developed and tested on debian 12 linux
should work fine on mac OS X and later, but not tested
currently uses "AVG" crack rate. Can be changed by modifying regex -- re := regexp.MustCompile...

example:
./crackmon ./hashcat {hashcat args} 			(defaults to -time 1m -crack 1)
./crackmon -t 2 -c 100 ./hashcat {hashcat args} 	(custom -time 2m -crack 100)

version:
v2023-10-01.1700; initial github release
v2023-10-02.1030-debug; fixed timeT; added more debugging info
v2023-10-04.1545-winpty; added pty support for windows; debug flag; changed CUR to AVG
v2023-10-07.1520-pty; refactored code, added logic for hashcat Paused, Running and Stopped status, added support for user sending commands to hashcat
*/

func help() {
	fmt.Fprint(os.Stderr, `Examples:

(Defaults to -time 1m -crack 1)
./crackmon ./hashcat {hashcat args}

Custom: -time 2m -crack 100
./crackmon -t 2 -c 100 ./hashcat {hashcat args}

All flags:
	-t      	time threshold in minutes
	-c      	average cracks threshold per -t
	-debug  	enable debug output
	-help   	show this help menu
	-version	show version info
`)
}

func version(debug bool) {
	fmt.Fprintln(os.Stderr, "crackmon v2023-10-07.1520-winpty")
	fmt.Fprintln(os.Stderr, "https://github.com/cyclone-github/crackmon")
	if debug {
		detectedOS := checkOS()
		fmt.Fprintf(os.Stderr, "\nDetected OS: %s\n", detectedOS)
	}
}

func cyclone() {
	codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
	codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
	fmt.Fprintln(os.Stderr, string(codedByDecoded))
}

func main() {
	timeT := flag.Int("t", 60, "Time threshold in seconds")
	crackT := flag.Int("c", 1, "Cracks per time threshold")
	debugFlag := flag.Bool("debug", false, "Enable debug mode")
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
		version(*debugFlag)
		os.Exit(0)
	}

	if *helpFlag {
		help()
		os.Exit(0)
	}

	// sanity checks
	cmdArgs := flag.Args()
	if len(cmdArgs) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing hashcat command to execute.\n")
		help()
		os.Exit(1)
	}
	cmdStr := strings.Join(cmdArgs, " ")

	if !strings.Contains(strings.ToLower(cmdStr), "hashcat") {
		fmt.Fprintln(os.Stderr, "Error: 'hashcat' must be part of the command.\n")
		help()
		os.Exit(1)
	}
	if !strings.Contains(cmdStr, "-o ") {
		fmt.Fprintln(os.Stderr, "hashcat outfile required. ex: -o founds.txt")
		os.Exit(1)
	}
	if strings.Contains(strings.ToLower(cmdStr), "--status-json") {
		fmt.Fprintln(os.Stderr, "\nWarning: --status-json is not allowed. Removing flag.")
		cmdStr = strings.ReplaceAll(cmdStr, "--status-json", "")
	}
	if !strings.Contains(strings.ToLower(cmdStr), "--status") {
		cmdStr += " --status"
	}
	if !strings.Contains(strings.ToLower(cmdStr), "--advice-disable") {
		cmdStr += " --advice-disable"
	}
	cmdStr = strings.Join(strings.Fields(cmdStr), " ")

	if *debugFlag {
		fmt.Fprintf(os.Stderr, "\nDEBUG:\tExecuting Command:\n%s\n\n", cmdStr)
		fmt.Fprintf(os.Stderr, "DEBUG:\t-t = %d\n", *timeT)
		fmt.Fprintf(os.Stderr, "DEBUG:\t-c = %d\n", *crackT)
		fmt.Fprintf(os.Stderr, "DEBUG:\tOS = %s\n\n", detectedOS)
	}

	re := regexp.MustCompile(`AVG:(\d+)\.`)

	initializeAndExecute(cmdStr, *timeT, *crackT, re, *debugFlag)
}
