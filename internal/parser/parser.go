package parser

import (
	"strings"

	"github.com/codecrafters-io/shell-starter-go/internal/types"
)

// ParseCommand parses a command string into a Command struct
func ParseCommand(input string) types.Command {
	tokens := tokenize(input)
	cmd := types.Command{}

	// Process tokens looking for redirection
	for i := 0; i < len(tokens); i++ {
		switch {
		// redirects stdout to a file (overwrite mode)
		case (tokens[i] == ">" || tokens[i] == "1>") && i+1 < len(tokens):
			cmd.OutputFile = tokens[i+1]
			cmd.AppendOutput = false
			i++

		// redirects stderr to a file (overwrite mode)
		case tokens[i] == "2>" && i+1 < len(tokens):
			cmd.ErrorFile = tokens[i+1]
			cmd.AppendOutput = false
			i++

		// redirects stdout to a file (append mode)
		case (tokens[i] == ">>" || tokens[i] == "1>>") && i+1 < len(tokens):
			cmd.OutputFile = tokens[i+1]
			cmd.AppendOutput = true
			i++

		// redirects stderr to a file (append mode)
		case tokens[i] == "2>>" && i+1 < len(tokens):
			cmd.ErrorFile = tokens[i+1]
			cmd.AppendOutput = true
			i++

		case cmd.Name == "":
			cmd.Name = tokens[i]
		default:
			cmd.Args = append(cmd.Args, tokens[i])
		}
	}

	return cmd
}

// tokenize splits the input into tokens
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
