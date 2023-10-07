//go:build windows
// +build windows

package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/UserExistsError/conpty"
)

// v2023-10-07.1520

var cptyInstance *conpty.ConPty

func initializeConPTY(fullCmd string) error {
	var err error
	cptyInstance, err = conpty.Start(fullCmd)
	if err != nil {
		return err
	}
	return nil
}

func windowsSendB(stdin io.Writer) {
	_, err := cptyInstance.Write([]byte("b\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send 'b' command: %v", err)
	}
}

func windowsSendQ(stdin io.Writer) {
	_, err := cptyInstance.Write([]byte("q\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send 'q' command: %v", err)
	}
	cptyInstance.Close()
	time.Sleep(1 * time.Second)
	os.Exit(0)
}

func initializeAndExecute(cmdStr string, timeT int, crackT int, re *regexp.Regexp, debug bool) {
	err := initializeConPTY(cmdStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing ConPTY:", err)
		return
	}
	sendB = windowsSendB
	sendQ = windowsSendQ

	// listen for user commands
	go ReadUserInput(cptyInstance)

	// initialize common logic
	initializeAndExecuteCommon(cmdStr, timeT, crackT, re, debug, cptyInstance, cptyInstance, checkOS)
}
