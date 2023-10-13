package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
)

/*
GNU GENERAL PUBLIC LICENSE Version 2, June 1991
https://github.com/cyclone-github/crackmon/blob/main/LICENSE

crackmon - by cyclone
original idea by: https://github.com/justpretending/avgdrop
hashcat wrapper tool to simulate pressing "b" key to bypass current hashcat attack if cracking rate goes below threshold
developed and tested on debian 12 linux

example:
./crackmon ./hashcat {hashcat args} 				(defaults to -time 1m -crack 1)
./crackmon -t 5 -c 100 ./hashcat {hashcat args}		(custom -time 5m -crack 100)

-t (time in minutes)	= minimum runtime in minutes
-c (total cracks)		= cumulative average cracks threshold

version:
v2023-10-01.1700; initial github release
v2023-10-02.1030-debug; fixed timeT; added more debugging info
v2023-10-04.1545-winpty; added pty support for windows; debug flag; changed CUR to AVG
v2023-10-07.1520-pty; refactored code, added logic for hashcat Paused, Running and Stopped status, added support for user sending commands to hashcat
v2023-10-13.1445; fixed https://github.com/cyclone-github/crackmon/issues/4; refactored sendX commands
*/

func help() {
	fmt.Fprint(os.Stderr, `Examples:

(Defaults to -time 1m -crack 1)
./crackmon ./hashcat {hashcat args}

Custom: -time 5m -crack 100
./crackmon -t 5 -c 100 ./hashcat {hashcat args}

All flags:
	-t      	minimum runtime in minutes
	-c      	cumulative average cracks threshold
	-debug  	enable debug output
	-help   	show this help menu
	-version	show version info
`)
}

// version func
func version(debug bool) {
	fmt.Fprintln(os.Stderr, "crackmon 2023-10-13.1445")
	fmt.Fprintln(os.Stderr, "https://github.com/cyclone-github/crackmon")
	if debug {
		detectedOS := checkOS()
		fmt.Fprintf(os.Stderr, "\nDetected OS: %s\n", detectedOS)
	}
}

// cyclone info
func cyclone() {
	codedBy := "Q29kZWQgYnkgY3ljbG9uZSA7KQo="
	codedByDecoded, _ := base64.StdEncoding.DecodeString(codedBy)
	fmt.Fprintln(os.Stderr, string(codedByDecoded))
}

// main func
func main() {
	timeT := flag.Int("t", 1, "Minimum runtime in minutes")
	crackT := flag.Int("c", 1, "Cumulative avg cracks threshold")
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

	// cli sanity checks
	cmdArgs := flag.Args()
	if len(cmdArgs) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing hashcat command to execute.\n")
		help()
		os.Exit(1)
	}
	if *timeT < 1 || *timeT > 100000 {
		fmt.Fprintf(os.Stderr, "Invalid value for -t. Must be between 1 and 100000.\n")
		os.Exit(1)
	}

	if *crackT < 1 || *crackT > 100000 {
		fmt.Fprintf(os.Stderr, "Invalid value for -c. Must be between 1 and 100000.\n")
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

	initializeAndExecute(cmdStr, *timeT, *crackT, *debugFlag)
}
