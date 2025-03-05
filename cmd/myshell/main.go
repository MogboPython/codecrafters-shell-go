package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

type Command struct {
	name         string
	args         []string
	outputFile   string
	appendOutput bool
	errorFile    string
}

var shellCommands = map[string]bool{
	"echo": true,
	"exit": true,
	"pwd":  true,
	"type": true,
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "\r$ ")

		input := readInput(os.Stdin)
		input = strings.TrimSpace(input)

		cmd := parseCommand(input)
		if err := executeCommand(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func readInput(rd io.Reader) (input string) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)
	r := bufio.NewReader(rd)

	for {
		c, _, err := r.ReadRune()
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch c {
		case '\x03': // Ctrl+C
			os.Exit(0)
		case '\r', '\n': // Enter
			fmt.Fprint(os.Stdout, "\r\n")
			return input
		case '\x7F': // Backspace
			if length := len(input); length > 0 {
				input = input[:length-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}
		case '\t': // Tab
			suffix := autocomplete(input)
			if suffix != "" {
				input += suffix + " "
				fmt.Fprint(os.Stdout, suffix+" ")
			} else {
				fmt.Fprint(os.Stdout, "\a")
			}
		default:
			input += string(c)
			fmt.Fprint(os.Stdout, string(c))
		}
	}
}

func autocomplete(prefix string) (suffix string) {
	if prefix == "" {
		return
	}
	suffixes := []string{}
	for cmd := range shellCommands {
		after, found := strings.CutPrefix(cmd, prefix)
		if found {
			suffixes = append(suffixes, after)
		}
	}
	if len(suffixes) == 1 {
		return suffixes[0]
	}
	return
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

	// Process tokens looking for redirection
	for i := 0; i < len(tokens); i++ {
		switch {
		case (tokens[i] == ">" || tokens[i] == "1>") && i+1 < len(tokens):
			cmd.outputFile = tokens[i+1]
			cmd.appendOutput = false
			i++ // Skip the filename token
		case tokens[i] == "2>" && i+1 < len(tokens):
			cmd.errorFile = tokens[i+1]
			cmd.appendOutput = false
			i++
		case (tokens[i] == ">>" || tokens[i] == "1>>") && i+1 < len(tokens):
			cmd.outputFile = tokens[i+1]
			cmd.appendOutput = true
			i++
		case tokens[i] == "2>>" && i+1 < len(tokens):
			cmd.errorFile = tokens[i+1]
			cmd.appendOutput = true
			i++
		case cmd.name == "":
			cmd.name = tokens[i]
		default:
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
		return executeExternalCommand(cmd)
	}
	return nil
}

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

func executeExternalCommand(cmd Command) error {
	execCmd := exec.Command(cmd.name, cmd.args...)

	var stdout, stderr io.Writer = os.Stdout, os.Stderr
	var outFile, errFile *os.File
	var err error

	// Set up output redirection
	if cmd.outputFile != "" {
		outFile = createOutputfile(cmd.outputFile, cmd.appendOutput)
		if outFile != nil {
			defer outFile.Close()
			stdout = outFile
		}
	}

	// Set up error redirection
	if cmd.errorFile != "" {
		errFile = createOutputfile(cmd.errorFile, cmd.appendOutput)
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
			return fmt.Errorf("%s: command not found", cmd.name)
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

	// Handle stdout redirection
	if cmd.outputFile != "" {
		file := createOutputfile(cmd.outputFile, cmd.appendOutput)
		if file == nil {
			return fmt.Errorf("failed to redirect output to %s", cmd.outputFile)
		}
		defer file.Close()
		os.Stdout = file
	}

	// Handle stderr redirection
	if cmd.errorFile != "" {
		errFile := createOutputfile(cmd.errorFile, cmd.appendOutput)
		if errFile == nil {
			return fmt.Errorf("failed to redirect error to %s", cmd.errorFile)
		}
		defer errFile.Close()
		os.Stderr = errFile
	}

	// Execute the command
	return execute()
}
