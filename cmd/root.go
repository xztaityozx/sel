package cmd

import (
	"bufio"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
	"xztaityozx/sel/option"
	"xztaityozx/sel/paser"
	"xztaityozx/sel/rw"
)

var rootCmd = &cobra.Command{
	Use:   "sel [OPTION] [QUERY]",
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
	Run: func(cmd *cobra.Command, queries []string) {
		opt := option.Option{
			InPlace: viper.GetBool("in-place"),
			Backup:  viper.GetBool("backup"),
			DelimiterOption: option.DelimiterOption{
				Input:  viper.GetString("input-delimiter"),
				OutPut: viper.GetString("output-delimiter"),
			},
			InputFiles: option.InputFiles{Files: viper.GetStringSlice("input-files")},
		}

		if opt.InPlace && len(opt.Files) == 0 {
			fatal("cannot in place stdin")
		}

		if err := do(opt, queries); err != nil {
			fatal(err)
		}

	},
}

func init() {
	rootCmd.Flags().BoolP("in-place", "i", false, "edit files in place")
	rootCmd.Flags().BoolP("backup", "b", false, "make backup when enable -i/--in-place option")
	rootCmd.Flags().StringSliceP("input-files", "f", nil, "input files")
	rootCmd.Flags().StringP("input-delimiter", "d", " ", "sets field delimiter(input)")
	rootCmd.Flags().StringP("output-delimiter", "D", " ", "sets field delimiter(output)")

	for _, key := range []string{
		"in-place", "backup",
		"input-files",
		"input-delimiter", "output-delimiter",
	} {
		_ = viper.BindPFlag(key, rootCmd.Flags().Lookup(key))
	}
}

func fatal(ifs ...interface{}) {
	_, _ = fmt.Fprint(os.Stderr, aurora.Red("Error:"))
	for _, v := range ifs {
		_, _ = fmt.Fprint(os.Stderr, v)
	}
	_, _ = fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

func do(opt option.Option, queries []string) error {

	pr, err := paser.New(queries...).Parse()
	if err != nil {
		return err
	}

	selector := func(s string) (string, error) {
		split, err := pr.Select(strings.Split(s, opt.Input))
		if err != nil {
			return "", err
		}

		return strings.Join(split, opt.OutPut), nil
	}

	if len(opt.Files) == 0 {
		// from stdin
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			fmt.Println(selector(scan.Text()))
		}
		_ = os.Stdout.Close()
		_ = os.Stdin.Close()
	} else {
		inputs, err := opt.Enumerate()
		if err != nil {
			return err
		}
		for _, v := range inputs {
			if err := rw.ReadWrite(v, opt.InPlace, opt.Backup, selector); err != nil {
				return err
			}
		}
	}

	return nil
}
