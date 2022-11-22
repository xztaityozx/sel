package cmd

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/xztaityozx/sel/iterator"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xztaityozx/sel/column"
	"github.com/xztaityozx/sel/option"
	"github.com/xztaityozx/sel/parser"
)

var rootCmd = &cobra.Command{
	Use:   "sel [queries...]",
	Short: "select column",
	Long: `
          _ 
 ___  ___| |
/ __|/ _ \ |
\__ \  __/ |
|___/\___|_|

__sel__ect column`,
	Args:    cobra.MinimumNArgs(1),
	Version: "1.1.7",
	Run: func(cmd *cobra.Command, args []string) {
		opt := option.NewOption(viper.GetViper())
		selectors, err := parser.Parse(args)
		if err != nil {
			log.Fatalln(err)
		}

		w := column.NewWriter(opt.OutPutDelimiter, os.Stdout)

		var iter iterator.IEnumerable

		// これから使うイテレーターを生成
		// オプションのON/OFFで行の分割戦略を変える
		if opt.UseRegexp {
			// Regexpを使う系の分割。遅め
			r, err := regexp.Compile(opt.InputDelimiter)
			if err != nil {
				log.Fatalln(err)
			}

			if opt.SplitBefore {
				// 事前に分割する。選択しないカラムも分割するが、後半のカラムを選択するときにはこちらが有利
				iter = iterator.NewPreSplitByRegexpIterator("", r, opt.RemoveEmpty)
			} else {
				// 欲しいところまで分割する。前の方に位置するカラムだけを選ぶ時に有利。
				// 負のインデックスを指定する場合は全部分割してしまうので不利
				iter = iterator.NewRegexpIterator("", r, opt.RemoveEmpty)
			}
		} else {
			if opt.SplitBefore {
				// 事前に分割する。regexp版と説明は同じ
				iter = iterator.NewPreSplitIterator("", opt.InputDelimiter, opt.RemoveEmpty)
			} else {
				// 最速。ただし、シンプルなIndex指定の時だけ
				iter = iterator.NewIterator("", opt.InputDelimiter, opt.RemoveEmpty)
			}
		}

		if len(opt.Files) != 0 {
			files, err := opt.InputFiles.Enumerate()
			if err != nil {
				log.Fatalln(err)
			}

			for _, file := range files {
				if fp, err := os.OpenFile(file, os.O_RDONLY, 0644); err != nil {
					log.Fatalln(err)
				} else {
					if err := run(fp, iter, w, selectors); err != nil {
						log.Fatalln(err)
					}
				}
			}
		} else {
			if err := run(os.Stdin, iter, w, selectors); err != nil {
				log.Fatalln(err)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.Flags().StringSliceP(option.NameInputFiles, "f", nil, "input files")
	rootCmd.Flags().StringP(option.NameInputDelimiter, "d", " ", "sets field delimiter(input)")
	rootCmd.Flags().StringP(option.NameOutPutDelimiter, "D", " ", "sets field delimiter(output)")
	rootCmd.Flags().BoolP(option.NameRemoveEmpty, "r", false, "remove empty sequence")
	rootCmd.Flags().BoolP(option.NameUseRegexp, "g", false, "use regular expressions for input delimiter")
	rootCmd.Flags().BoolP(option.NameSplitBefore, "S", false, "split all column before select")
	rootCmd.Flags().BoolP(option.NameFieldSplit, "a", false, "Shorthand for -gd '\\s+'")
	_ = rootCmd.MarkFlagFilename(option.NameInputFiles)

	for _, key := range option.GetOptionNames() {
		_ = viper.BindPFlag(key, rootCmd.Flags().Lookup(key))
	}

	examples := []string{
		"",
		"$ cat /path/to/file | sel 1",
		"$ sel 1:10 -f ./file",
		"$ cat /path/to/file.csv | sel -d, 1 2 3 4 -1 -2 -3 -4",
		"$ sel 2:: -f ./file",
	}

	rootCmd.Example = strings.Join(examples, "\n\t")

	rootCmd.SetUsageTemplate(`Usage:
	{{.CommandPath}} [queries...]

Query:
	index                        select 'index'
	start:stop                   select columns from 'start' to 'stop'
	start:stop:step              select columns each 'step' from 'start' to 'stop'

	start:/end regexp/           select columns from 'start' to /end regexp/
	/start regexp/:end           select columns from /start regexp/ to 'end'
	/start regexp/:/end regexp/  select columns from /start regexp/ to /end regexp/

Examples:
{{.Example}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

func run(input *os.File, iter iterator.IEnumerable, writer *column.Writer, selectors []column.Selector) error {
	defer func(input *os.File) {
		err := input.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(input)

	scan := bufio.NewScanner(input)
	for scan.Scan() {
		iter.Reset(scan.Text())
		for _, selector := range selectors {
			err := selector.Select(writer, iter)
			if err != nil {
				return err
			}
		}

		if err := writer.WriteNewLine(); err != nil {
			return err
		}
	}

	return writer.Flush()
}
