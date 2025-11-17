//go:build linux
// +build linux

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty" // used for sending user keyboard commands to hashcat/mdxfind
)

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
	ptmx, err := pty.Start(cmd) // start command with pty
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting command with PTY:", err)
		return
	}
	defer func() { _ = ptmx.Close() }() // close pty

	// Runner-specific sendB / sendQ behavior
	switch currentRunner {
	case RunnerHashcat:
		// hashcat: "b" to bypass, "q" to quit
		sendB = func(stdin io.Writer) { linuxSendCmd("b", stdin) }
		sendQ = func(stdin io.Writer) { linuxSendCmd("q", stdin) }
	case RunnerMDXFind:
		// mdxfind: no bypass/quit keys; use Ctrl+C for both
		sendB = func(stdin io.Writer) { linuxSendCmd("\x03", stdin) } // Ctrl+C
		sendQ = func(stdin io.Writer) { linuxSendCmd("\x03", stdin) } // Ctrl+C
	default:
		// fail-safe: use Ctrl+C if Runner unknown
		sendB = func(stdin io.Writer) { linuxSendCmd("\x03", stdin) }
		sendQ = func(stdin io.Writer) { linuxSendCmd("\x03", stdin) }
	}

	// listen for user commands
	go ReadUserInput(ptmx)

	// initialize common logic
	initializeAndExecuteCommon(timeT, crackT, debug, ptmx, ptmx)
}
