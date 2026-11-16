package commands

import "os"

// IsTerminalFunc reports whether stdin is an interactive terminal.
type IsTerminalFunc func() bool

// StdinIsTerminal is the real implementation: stdin is a character device.
func StdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return fi.Mode()&os.ModeCharDevice != 0
}
