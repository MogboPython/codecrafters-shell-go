package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	var command string
	var args string
	var shellCommands = map[string]bool{
		"echo": true,
		"exit": true,
		"pwd":  true,
		"type": true,
	}

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("An error occured: ", err)
			os.Exit(1)
		}

		// TODO: use this
		// input = strings.Trim(input, "\r\n")
		// args := strings.Split(input, " ")
		// command, args := args[0], args[1:]

		input = strings.TrimSpace(input)
		parts := strings.SplitN(input, " ", 2)
		command = parts[0]
		args = ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(args)
		case "type":
			if shellCommands[args] {
				fmt.Println(args + " is a shell builtin")
			} else if path, err := exec.LookPath(args); err == nil {
				fmt.Println(args + " is " + path)
			} else {
				fmt.Println(args + ": not found")
			}

		case "pwd":
			wd := getWorkingDirectory()
			fmt.Println(wd)

		case "cd":
			err := changeWorkingDirectory(args)
			if err != nil {
				fmt.Printf("%s: %s: No such file or directory\n", command, args)
			}

		default:
			cmd := exec.Command(command, args)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				fmt.Printf("%s: command not found\n", command)
			}
		}
	}
}

func getWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return ""
	}
	return wd
}

func changeWorkingDirectory(path string) error {
	err := os.Chdir(path)
	if err != nil {
		return fmt.Errorf("error changing working directory: %v", err)
	}
	return nil
}
