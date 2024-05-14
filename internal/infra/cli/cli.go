package cli

import (
	"github.com/spf13/cobra"
)

type CLIInterface interface {
	Start() error
}

type CLI struct {
	RootCmd *cobra.Command
}

func NewCLI(rootCmd *cobra.Command) *CLI {
	return &CLI{
		RootCmd: rootCmd,
	}
}

func (c *CLI) Start() error {
	return c.RootCmd.Execute()
}
