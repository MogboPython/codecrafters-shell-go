package main

import (
	"bufio"
	"fmt"
	"os"
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
		parts := strings.SplitN(input, " ", 2)
		command = parts[0]

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			args = parts[1]
			fmt.Println(args)
		case "type":
			args = parts[1]
			if shellCommands[args] {
				fmt.Println(args + " is a shell builtin")
			} else {
				fmt.Println(args + ": not found")
			}

		default:
			fmt.Println(command + ": command not found")
		}
	}
}
