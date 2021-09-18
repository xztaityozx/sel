package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate completion script",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			err := cmd.Root().GenBashCompletion(os.Stdout)
			if err != nil {
				log.Fatalln(err)
			}
		case "zsh":
			err := cmd.Root().GenZshCompletion(os.Stdout)
			if err != nil {
				log.Fatalln(err)
			}
		case "fish":
			err := cmd.Root().GenFishCompletion(os.Stdout, true)
			if err != nil {
				log.Fatalln(err)
			}
		case "powershell":
			err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			if err != nil {
				log.Fatalln(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
