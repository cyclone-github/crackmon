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

// sendX func (used for hashcat commands "b" / "q")
func windowsSendCmd(cmd string) {
	_, err := cptyInstance.Write([]byte(cmd + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send '%s' command: %v", cmd, err)
	}
}

// helper for sending raw bytes (used for Ctrl+C to mdxfind)
func windowsSendRaw(data []byte) {
	_, err := cptyInstance.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send raw data: %v", err)
	}
}

// initialize OS specific logic
func initializeAndExecute(cmdStr string, timeT int, crackT int, debug bool) {
	err := initializeConPTY(cmdStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing ConPTY:", err)
		return
	}

	switch currentRunner {
	case RunnerHashcat:
		// hashcat: "b" to bypass, "q" to quit
		sendB = func(stdin io.Writer) { windowsSendCmd("b") }
		sendQ = func(stdin io.Writer) {
			windowsSendCmd("q")
			cptyInstance.Close()
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}
	case RunnerMDXFind:
		// mdxfind: no bypass/quit keys; use Ctrl+C / cpty close instead
		sendB = func(stdin io.Writer) {
			fmt.Println("Sending quit signal...")
			windowsSendRaw([]byte{0x03}) // Ctrl+C doesn't work as expected on windows
			cptyInstance.Close()         // close cpty
			time.Sleep(1 * time.Second)
			os.Exit(0) // force exit
		}
		sendQ = func(stdin io.Writer) {
			fmt.Println("Sending quit signal...")
			windowsSendRaw([]byte{0x03}) // Ctrl+C doesn't work as expected on windows
			cptyInstance.Close()         // close cpty
			time.Sleep(1 * time.Second)
			os.Exit(0) // force exit
		}
	default:
		// fail-safe: use Ctrl+C
		sendB = func(_ io.Writer) {
			fmt.Println("Sending quit signal...")
			windowsSendRaw([]byte{0x03}) // Ctrl+C doesn't work as expected on windows
			cptyInstance.Close()         // close cpty
			time.Sleep(1 * time.Second)
			os.Exit(0) // force exit
		}
		sendQ = func(_ io.Writer) {
			fmt.Println("Sending quit signal...")
			windowsSendRaw([]byte{0x03}) // Ctrl+C doesn't work as expected on windows
			cptyInstance.Close()         // close cpty
			time.Sleep(1 * time.Second)
			os.Exit(0) // force exit
		}
	}

	// listen for user commands
	go ReadUserInput(cptyInstance)

	// initialize common logic
	initializeAndExecuteCommon(timeT, crackT, debug, cptyInstance, cptyInstance)
}
