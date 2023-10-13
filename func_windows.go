//go:build windows
// +build windows

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/UserExistsError/conpty"
)

/*
v2023-10-07.1520
v2023-10-13.1445; refactored sendX commands
*/

var cptyInstance *conpty.ConPty

// conpty func for windows pty support
func initializeConPTY(fullCmd string) error {
	var err error
	cptyInstance, err = conpty.Start(fullCmd)
	if err != nil {
		return err
	}
	return nil
}

// sendX func
func windowsSendCmd(cmd string, stdin io.Writer) {
	_, err := cptyInstance.Write([]byte(cmd + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send '%s' command: %v", cmd, err)
	}
}

// initialize OS specific logic
func initializeAndExecute(cmdStr string, timeT int, crackT int, debug bool) {
	err := initializeConPTY(cmdStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing ConPTY:", err)
		return
	}
	sendB = func(stdin io.Writer) { windowsSendCmd("b", stdin) }
	sendQ = func(stdin io.Writer) {
		windowsSendCmd("q", stdin)
		cptyInstance.Close()
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}

	// listen for user commands
	go ReadUserInput(cptyInstance)

	// initialize common logic
	initializeAndExecuteCommon(cmdStr, timeT, crackT, debug, cptyInstance, cptyInstance, checkOS)
}
