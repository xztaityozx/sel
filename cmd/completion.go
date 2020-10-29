package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type Completion struct {
	W io.Writer
}

// Gen generate completion script for shell
func (c Completion) Gen(command *cobra.Command, shell string) error {
	if shell == "bash" {
		return command.GenBashCompletion(c.W)
	} else if shell == "zsh" {
		return command.GenZshCompletion(c.W)
	} else if shell == "fish" {
		return command.GenFishCompletion(c.W, true)
	} else if shell == "pwsh" || shell == "PowerShell" {
		return command.GenPowerShellCompletion(c.W)
	} else {
		return fmt.Errorf("failed to generate completion for %s", shell)
	}
}
