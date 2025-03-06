package types

type Command struct {
	Name         string
	Args         []string
	OutputFile   string
	AppendOutput bool
	ErrorFile    string
}

type AutoCompleteResult int

const (
	AUTOCOMPLETE_ERROR = iota
	AUTOCOMPLETE_DIRECT_MATCH
	AUTOCOMPLETE_MULTI_MATCH
	AUTOCOMPLETE_NO_MATCH
)

// built-in shell commands
var ShellCommands = map[string]bool{
	"echo": true,
	"exit": true,
	"pwd":  true,
	"type": true,
}
