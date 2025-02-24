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
	var args []string
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

		input = strings.TrimSpace(input)

		parsed := parseCommand(input)
		if len(parsed) == 0 {
			continue
		}

		command = parsed[0]
		args = parsed[1:]

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "type":
			if shellCommands[args[0]] {
				fmt.Println(args[0] + " is a shell builtin")
			} else if path, err := exec.LookPath(args[0]); err == nil {
				fmt.Println(args[0] + " is " + path)
			} else {
				fmt.Println(args[0] + ": not found")
			}

		case "pwd":
			wd := getWorkingDirectory()
			fmt.Println(wd)

		case "cd":
			err := changeWorkingDirectory(args[0])
			if err != nil {
				fmt.Printf("%s: %s: No such file or directory\n", command, args)
			}

		default:
			cmd := exec.Command(command, args[0])
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
	if path == "" || path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %v", err)
		}
		return os.Chdir(homeDir)
	}

	// Handle paths starting with "~/"
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %v", err)
		}
		path = homeDir + path[1:] // Replace ~ with home directory
	}

	return os.Chdir(path)
}

// parseCommand splits the input into tokens, respecting single quotes
func parseCommand(input string) []string {
	var tokens []string
	var currentText strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		switch char {
		case '\'':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if currentText.Len() > 0 {
					tokens = append(tokens, currentText.String())
					currentText.Reset()
				}
			} else {
				currentText.WriteByte(char)
			}
		default:
			currentText.WriteByte(char)
		}
	}

	if currentText.Len() > 0 {
		tokens = append(tokens, currentText.String())
	}

	return tokens
}
