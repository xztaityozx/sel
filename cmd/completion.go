package cmd

import (
  "os"

  "github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
  Use:                   "completion [bash|zsh|fish|powershell]",
  Short:                 "Generate completion script",
  DisableFlagsInUseLine: true,
  ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
  Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
  RunE: func(cmd *cobra.Command, args []string) error {
    cmd.SilenceUsage = true
    switch args[0] {
    case "bash":
      return cmd.Root().GenBashCompletion(os.Stdout)
    case "zsh":
      return cmd.Root().GenZshCompletion(os.Stdout)
    case "fish":
      return cmd.Root().GenFishCompletion(os.Stdout, true)
    case "powershell":
      return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
    }
    return nil
  },
}

func init() {
  rootCmd.AddCommand(completionCmd)
  completionCmd.SetUsageTemplate(completionCmd.Use)
}
