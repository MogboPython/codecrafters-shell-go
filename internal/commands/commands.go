package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/internal/types"
)

// ExecuteCommand executes a command
func ExecuteCommand(cmd types.Command) error {
	switch cmd.Name {
	case "cd":
		err := changeWorkingDirectory(cmd.Args[0])
		if err != nil {
			fmt.Printf("%s: %s: No such file or directory\n", cmd.Name, cmd.Args[0])
		}

	case "pwd":
		wd := getWorkingDirectory()
		fmt.Println(wd)

	case "exit":
		os.Exit(0)

	case "echo":
		return executeWithRedirection(cmd, func() error {
			fmt.Println(strings.Join(cmd.Args, " "))
			return nil
		})

	case "type":
		if types.ShellCommands[cmd.Args[0]] {
			fmt.Println(cmd.Args[0] + " is a shell builtin")
		} else if path, err := exec.LookPath(cmd.Args[0]); err == nil {
			fmt.Println(cmd.Args[0] + " is " + path)
		} else {
			fmt.Println(cmd.Args[0] + ": not found")
		}

	default:
		return executeExternalCommand(cmd)
	}
	return nil
}
