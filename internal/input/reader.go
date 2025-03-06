package input

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/shell-starter-go/internal/autocomplete"
	"github.com/codecrafters-io/shell-starter-go/internal/types"

	"golang.org/x/term"
)

// ReadInput reads user input and supports autocomplete
func ReadInput(rd io.Reader) (input string) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	r := bufio.NewReader(rd)
	tabPresses := 0

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
			tabPresses = 0

		case '\t': // Tab
			tabPresses++
			match, AutoCompleteResult := autocomplete.Complete(input)
			switch AutoCompleteResult {
			case types.AUTOCOMPLETE_DIRECT_MATCH:
				input += match
				fmt.Fprint(os.Stdout, match)
				tabPresses = 0
			case types.AUTOCOMPLETE_NO_MATCH:
				fmt.Fprint(os.Stdout, "\a")
				tabPresses = 0
			case types.AUTOCOMPLETE_MULTI_MATCH:
				if tabPresses == 1 {
					fmt.Fprint(os.Stdout, "\a")
				} else {
					term.Restore(fd, oldState)
					fmt.Fprintf(os.Stdout, "\r\n%s\r\n", match)
					fmt.Fprint(os.Stdout, "$ "+input)
					newState, _ := term.MakeRaw(fd)
					oldState = newState
					tabPresses = 0
				}
			}

		default:
			input += string(c)
			fmt.Fprint(os.Stdout, string(c))
			tabPresses = 0
		}
	}
}
