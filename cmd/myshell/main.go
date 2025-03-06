package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/internal/commands"
	"github.com/codecrafters-io/shell-starter-go/internal/input"
	"github.com/codecrafters-io/shell-starter-go/internal/parser"
)

func main() {
	for {
		fmt.Fprint(os.Stdout, "\r$ ")

		input := input.ReadInput(os.Stdin)
		input = strings.TrimSpace(input)

		cmd := parser.ParseCommand(input)
		if err := commands.ExecuteCommand(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
