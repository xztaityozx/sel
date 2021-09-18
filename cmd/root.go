package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xztaityozx/sel/option"
)

var rootCmd = &cobra.Command{
	Use:   "sel",
	Short: "select column",
	Long: `
          _ 
 ___  ___| |
/ __|/ _ \ |
\__ \  __/ |
|___/\___|_|

__sel__ect column`,
	Example: "sel 1",
	Args:    cobra.MinimumNArgs(1),
	Version: "1.1.1",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.Flags().StringSliceP(option.NameInputFiles, "f", nil, "input files")
	rootCmd.Flags().StringP(option.NameInputDelimiter, "d", " ", "sets field delimiter(input)")
	rootCmd.Flags().StringP(option.NameOutPutDelimiter, "D", " ", "sets field delimiter(output)")
	rootCmd.Flags().BoolP(option.NameRemoveEmpty, "r", false, "remove empty sequence")
	rootCmd.Flags().BoolP(option.NameUseRegexp, "g", false, "use regular expressions for input delimiter")
	_ = rootCmd.MarkFlagFilename(option.NameInputFiles)

	for _, key := range option.GetOptionNames() {
		_ = viper.BindPFlag(key, rootCmd.Flags().Lookup(key))
	}
}
