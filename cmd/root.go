package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xztaityozx/sel/option"
	"github.com/xztaityozx/sel/parser"
	"github.com/xztaityozx/sel/rw"
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
	Version: "1.1.0",
	PreRun: func(cmd *cobra.Command, args []string) {
		if shell := viper.GetString("completion"); len(shell) != 0 {
			err := Completion{W: os.Stdout}.Gen(cmd, shell)
			if err != nil {
				fatal(err)
			}

			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, queries []string) {
		opt := option.Option{
			InPlace: viper.GetBool("in-place"),
			Backup:  viper.GetBool("backup"),
			DelimiterOption: option.DelimiterOption{
				InputDelimiter:  viper.GetString("input-delimiter"),
				OutPutDelimiter: viper.GetString("output-delimiter"),
				RemoveEmpty:     viper.GetBool("remove-empty"),
				UseRegexp:       viper.GetBool("use-regexp"),
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
	rootCmd.Flags().BoolP("remove-empty", "r", false, "remove empty sequence")
	rootCmd.Flags().String("completion", "", "generate completion")
	rootCmd.Flags().BoolP("use-regexp", "g", false, "use regular expressions for input delimiter")
	_ = rootCmd.Flags().MarkHidden("completion")
	_ = rootCmd.MarkFlagFilename("input-files")

	for _, key := range []string{
		"in-place", "backup",
		"input-files",
		"input-delimiter", "output-delimiter", "use-regexp",
		"completion",
		"remove-empty",
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

	pr, err := parser.New(queries...).Parse()
	if err != nil {
		return err
	}

	// --use-regexpならRegexpをコンパイルしておく
	var delmRegexp *regexp.Regexp
	if opt.UseRegexp {
		var err error
		delmRegexp, err = regexp.Compile(opt.InputDelimiter)
		if err != nil {
			return err
		}
	}

	// 文字列の分割Function
	spliter := func(in string) []string {
		if opt.UseRegexp {
			return delmRegexp.Split(in, -1)
		} else {
			return strings.Split(in, opt.InputDelimiter)
		}
	}

	// カラム選択Function
	selector := func(s string) (string, error) {
		line := func() []string {
			var rt []string
			for _, v := range spliter(s) {
				if opt.RemoveEmpty && len(v) == 0 {
					continue
				}
				rt = append(rt, v)
			}
			return rt
		}()
		split, err := pr.Select(line)
		if err != nil {
			return "", err
		}

		return strings.Join(split, opt.OutPutDelimiter), nil
	}

	if len(opt.Files) == 0 {
		// from stdin
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			line, err := selector(scan.Text())
			if err != nil {
				return err
			}
			fmt.Println(line)
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
