package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/codecrafters-io/shell-starter-go/internal/types"
)

// executeExternalCommand executes a non-builtin command
func executeExternalCommand(cmd types.Command) error {
	execCmd := exec.Command(cmd.Name, cmd.Args...)

	var stdout, stderr io.Writer = os.Stdout, os.Stderr
	var outFile, errFile *os.File
	var err error

	// Set up output redirection
	if cmd.OutputFile != "" {
		outFile = createOutputfile(cmd.OutputFile, cmd.AppendOutput)
		if outFile != nil {
			defer outFile.Close()
			stdout = outFile
		}
	}

	// Set up error redirection
	if cmd.ErrorFile != "" {
		errFile = createOutputfile(cmd.ErrorFile, cmd.AppendOutput)
		if errFile != nil {
			defer errFile.Close()
			stderr = errFile
		}
	}

	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	// Run the command
	err = execCmd.Run()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return fmt.Errorf("%s: command not found", cmd.Name)
		}
	}

	return nil
}

// createOutputfile creates or opens a file for output redirection
func createOutputfile(fileName string, appendFlag bool) *os.File {
	flag := os.O_WRONLY | os.O_CREATE
	if appendFlag {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	outFile, err := os.OpenFile(fileName, flag, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return outFile
}

// executeWithRedirection executes the echo function with output redirection
func executeWithRedirection(cmd types.Command, execute func() error) error {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	cleanup := func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}

	defer cleanup()

	// Handle stdout redirection
	if cmd.OutputFile != "" {
		file := createOutputfile(cmd.OutputFile, cmd.AppendOutput)
		if file == nil {
			return fmt.Errorf("failed to redirect output to %s", cmd.OutputFile)
		}
		defer file.Close()
		os.Stdout = file
	}

	// Handle stderr redirection
	if cmd.ErrorFile != "" {
		errFile := createOutputfile(cmd.ErrorFile, cmd.AppendOutput)
		if errFile == nil {
			return fmt.Errorf("failed to redirect error to %s", cmd.ErrorFile)
		}
		defer errFile.Close()
		os.Stderr = errFile
	}

	// Execute the command
	return execute()
}
