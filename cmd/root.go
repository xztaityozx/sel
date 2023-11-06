package cmd

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"

	"github.com/xztaityozx/sel/iterator"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xztaityozx/sel/column"
	"github.com/xztaityozx/sel/option"
	"github.com/xztaityozx/sel/parser"
)

var Version string = "undefined"

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
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		opt := option.NewOption(viper.GetViper())
		selectors, err := parser.Parse(args)
		if err != nil {
			log.Fatalln(err)
		}

		w := column.NewWriter(opt.OutPutDelimiter, os.Stdout)

		if len(opt.Files) != 0 {
			files, err := opt.InputFiles.Enumerate()
			if err != nil {
				log.Fatalln(err)
			}

			for _, file := range files {
				if fp, err := os.Open(file); err != nil {
					log.Fatalln(err)
				} else {
					if err := run(fp, opt, w, selectors); err != nil {
						log.Fatalln(err)
					}
				}
			}
		} else {
			if err := run(os.Stdin, opt, w, selectors); err != nil {
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
	rootCmd.Flags().BoolP(option.NameFieldSplit, "a", false, "shorthand for -gd '\\s+'")
	rootCmd.Flags().Bool(option.NameCsv, false, "parse input file as CSV")
	rootCmd.Flags().Bool(option.NameTsv, false, "parse input file as TSV")
	_ = rootCmd.MarkFlagFilename(option.NameInputFiles)
	rootCmd.MarkFlagsMutuallyExclusive(option.NameCsv, option.NameTsv)

	for _, key := range option.GetOptionNames() {
		_ = viper.BindPFlag(key, rootCmd.Flags().Lookup(key))
	}

	examples := []string{
		"",
		"$ cat /path/to/file | sel 1",
		"$ sel 1:10 -f ./file",
		"$ cat /path/to/file.csv | sel -d, 1 2 3 4 -- -1 -2 -3 -4",
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

// run はあるファイルについて column.Selector によるカラム選択と column.Writer による書き出しを行う。ファイルはCloseされる
func run(input *os.File, option option.Option, w *column.Writer, selectors []column.Selector) error {
	defer func(input *os.File) {
		if err := input.Close(); err != nil {
			log.Fatalln(err)
		}
	}(input)

	iter, err := iterator.NewIEnumerable(option)
	if err != nil {
		return err
	}

	if ok, comma := option.IsXsv(); ok {
		r := csv.NewReader(input)
		r.Comma = comma

		var record []string
		var csvReadError error
		for {
			record, csvReadError = r.Read()
			if csvReadError != nil && csvReadError != io.EOF {
				return csvReadError
			}
			if csvReadError == io.EOF {
				break
			}

			iter.ResetFromArray(record)

			if err := selectAll(&iter, w, selectors); err != nil {
				return err
			}
		}

		return w.Flush()
	}

	scan := bufio.NewScanner(input)
	for scan.Scan() {
		iter.Reset(scan.Text())
		if err := selectAll(&iter, w, selectors); err != nil {
			return err
		}
	}

	return w.Flush()
}

func selectAll(iter *iterator.IEnumerable, w *column.Writer, selectors []column.Selector) error {
	for _, selector := range selectors {
		err := selector.Select(w, *iter)
		if err != nil {
			return err
		}
	}
	return w.WriteNewLine()
}
