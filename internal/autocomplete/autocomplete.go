package autocomplete

import (
	"os"
	"sort"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/internal/types"
)

// Cache of executables in PATH to avoid repeated filesystem lookups
var executableCache map[string]bool

func init() {
	// Initialize the cache
	executableCache = findExecutablesInPath()
}

// Complete attempts to autocomplete the given prefix
func Complete(prefix string) (string, types.AutoCompleteResult) {
	if prefix == "" {
		return "", types.AUTOCOMPLETE_NO_MATCH
	}

	var matches []string
	for cmd := range executableCache {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 0 {
		return "", types.AUTOCOMPLETE_NO_MATCH
	}

	if len(matches) == 1 {
		return matches[0][len(prefix):] + " ", types.AUTOCOMPLETE_DIRECT_MATCH
	}

	sort.Strings(matches)

	first, last := matches[0], matches[len(matches)-1]
	i := len(prefix)
	for i < len(first) && i < len(last) && first[i] == last[i] {
		i++
	}

	commonPrefix := first[:i]
	if len(commonPrefix) > len(prefix) {
		return commonPrefix[len(prefix):], types.AUTOCOMPLETE_DIRECT_MATCH
	}

	return strings.Join(matches, "  "), types.AUTOCOMPLETE_MULTI_MATCH
}

// findExecutablesInPath finds all executables in the PATH
func findExecutablesInPath() map[string]bool {
	result := make(map[string]bool)

	// Get PATH environment variable
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, ":")

	for cmd := range types.ShellCommands {
		result[cmd] = true
	}

	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				result[file.Name()] = true
			}
		}
	}

	return result
}
