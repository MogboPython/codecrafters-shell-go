package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	name       string
	args       []string
	outputFile string
	errorFile  string
}

var shellCommands = map[string]bool{
	"echo": true,
	"exit": true,
	"pwd":  true,
	"type": true,
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("An error occured: ", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		cmd := parseCommand(input)
		if err := executeCommand(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
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

func parseCommand(input string) Command {
	tokens := tokenize(input)
	cmd := Command{}

	// TODO: change to switch case
	// Process tokens looking for redirection
	for i := 0; i < len(tokens); i++ {
		if (tokens[i] == ">" || tokens[i] == "1>") && i+1 < len(tokens) {
			cmd.outputFile = tokens[i+1]
			// Skip the next token
			i++
		} else if tokens[i] == "2>" && i+1 < len(tokens) {
			cmd.errorFile = tokens[i+1]
			i++
		} else if cmd.name == "" {
			cmd.name = tokens[i]
		} else {
			cmd.args = append(cmd.args, tokens[i])
		}
	}

	return cmd
}

func executeCommand(cmd Command) error {
	switch cmd.name {
	case "cd":
		err := changeWorkingDirectory(cmd.args[0])
		if err != nil {
			fmt.Printf("%s: %s: No such file or directory\n", cmd.name, cmd.args[0])
		}

	case "pwd":
		wd := getWorkingDirectory()
		fmt.Println(wd)

	case "exit":
		os.Exit(0)

	case "echo":
		return executeWithRedirection(cmd, func() error {
			fmt.Println(strings.Join(cmd.args, " "))
			return nil
		})

	case "type":
		if shellCommands[cmd.args[0]] {
			fmt.Println(cmd.args[0] + " is a shell builtin")
		} else if path, err := exec.LookPath(cmd.args[0]); err == nil {
			fmt.Println(cmd.args[0] + " is " + path)
		} else {
			fmt.Println(cmd.args[0] + ": not found")
		}

	default:
		execCmd := exec.Command(cmd.name, cmd.args...)
		err := executeWithRedirection(cmd, func() error {
			execCmd.Stderr = os.Stderr
			execCmd.Stdout = os.Stdout
			return execCmd.Run()
		})
		if err != nil {
			fmt.Printf("%s: command not found\n", cmd.name)
		}
	}
	return nil
}

func executeWithRedirection(cmd Command, execute func() error) error {
	// Save original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	cleanup := func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}
	defer cleanup()

	if cmd.outputFile != "" {
		file, err := os.Create(cmd.outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file %s: %w", cmd.outputFile, err)
		}
		defer file.Close()
		os.Stdout = file
	}

	if cmd.errorFile != "" {
		errFile, err := os.Create(cmd.errorFile)
		if err != nil {
			// os.Stdout = oldStdout
			return fmt.Errorf("error creating error file %s: %w", cmd.errorFile, err)
		}
		defer errFile.Close()
		os.Stderr = errFile
	}

	// Execute the command
	err := execute()

	// Handle execution error
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Only print stderr if we're not already redirecting it
			if cmd.errorFile == "" {
				fmt.Fprintf(oldStderr, "%s", string(exitErr.Stderr))
			}
			return exitErr
		}
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}

// splits the input into tokens
func tokenize(input string) []string {
	var tokens []string
	var currentText strings.Builder
	var inSingleQuotes bool
	var inDoubleQuotes bool

	for i := 0; i < len(input); i++ {
		char := input[i]

		switch {
		case char == '\'':
			if !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
			} else {
				currentText.WriteByte(char)
			}

		case char == '"':
			if !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
			} else {
				currentText.WriteByte(char)
			}

		case char == '\\' && !inSingleQuotes:
			if i+1 < len(input) {
				// Handle escaped character
				nextChar := input[i+1]
				if inDoubleQuotes {
					// In double quotes, only escape \ and "
					switch nextChar {
					case '\\', '"':
						currentText.WriteByte(nextChar)
						i++ // Skip the next character
					default:
						currentText.WriteByte(char)
					}
				} else {
					// Outside quotes, escape any character
					currentText.WriteByte(nextChar)
					i++
				}
			} else {
				// Backslash at end of input
				currentText.WriteByte(char)
			}

		case char == ' ':
			if !inSingleQuotes && !inDoubleQuotes {
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
