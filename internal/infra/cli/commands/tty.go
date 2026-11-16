package commands

import (
	"os"

	"github.com/charmbracelet/x/term"
)

// IsTerminalFunc reports whether stdin is an interactive terminal.
type IsTerminalFunc func() bool

// StdinIsTerminal is the real implementation. It uses a real isatty check so
// that /dev/null (also a character device) is not mistaken for a terminal.
func StdinIsTerminal() bool {
	return term.IsTerminal(os.Stdin.Fd())
}
