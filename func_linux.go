//go:build linux || darwin
// +build linux darwin

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/creack/pty"
)

// v2023-10-07.1520

func linuxSendB(stdin io.Writer) {
	io.WriteString(stdin, "b")
}

func linuxSendQ(stdin io.Writer) {
	io.WriteString(stdin, "q")
}

func initializeAndExecute(cmdStr string, timeT int, crackT int, re *regexp.Regexp, debug bool) {
	cmdSlice := strings.Fields(cmdStr)
	cmdName := cmdSlice[0]
	cmdArgs := cmdSlice[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	ptmx, err := pty.Start(cmd) // start hashcat command with pty
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting command with PTY:", err)
		return
	}
	defer func() { _ = ptmx.Close() }() // close pty

	sendB = linuxSendB
	sendQ = linuxSendQ

	// listen for user commands
	go ReadUserInput(ptmx)

	// initialize common logic
	initializeAndExecuteCommon(cmdStr, timeT, crackT, re, debug, ptmx, ptmx, checkOS)
}
