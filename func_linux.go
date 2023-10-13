//go:build linux
// +build linux

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty" // used for sending user keyboard commands to hashcat
)

/*
v2023-10-07.1520
v2023-10-13.1445; refactored sendX commands
*/

// sendX func
func linuxSendCmd(cmd string, stdin io.Writer) {
	io.WriteString(stdin, cmd)
}

// initialize OS specific logic
func initializeAndExecute(cmdStr string, timeT int, crackT int, debug bool) {
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

	sendB = func(stdin io.Writer) { linuxSendCmd("b", stdin) }
	sendQ = func(stdin io.Writer) { linuxSendCmd("q", stdin) }

	// listen for user commands
	go ReadUserInput(ptmx)

	// initialize common logic
	initializeAndExecuteCommon(cmdStr, timeT, crackT, debug, ptmx, ptmx, checkOS)
}
