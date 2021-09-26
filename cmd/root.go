package cmd

import (
	"bufio"
	"log"
	"os"
	"strings"

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
	Version: "1.1.1",
	Run: func(cmd *cobra.Command, args []string) {
		opt := option.NewOption(viper.GetViper())
		selectors, err := parser.Parse(args)
		if err != nil {
			log.Fatalln(err)
		}

		w := column.NewWriter(opt.OutPutDelimiter, os.Stdout)
		var splitter column.Splitter
		if opt.UseRegexp {
			splitter, err = column.NewSplitterRegexp(opt.InputDelimiter, opt.RemoveEmpty)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			splitter = column.NewSplitter(opt.InputDelimiter, opt.RemoveEmpty)
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
					if err := run(fp, splitter, w, selectors); err != nil {
						log.Fatalln(err)
					}
				}
			}
		} else {
			if err := run(os.Stdin, splitter, w, selectors); err != nil {
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

func run(input *os.File, splitter column.Splitter, writer *column.Writer, selectors []column.Selector) error {
	defer func(input *os.File) {
		err := input.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(input)

	scan := bufio.NewScanner(input)
	for scan.Scan() {
		line := splitter.Split(scan.Text())
		for _, selector := range selectors {
			cols, err := selector.Select(line)
			if err != nil {
				return err
			}

			if err := writer.Write(cols); err != nil {
				return err
			}
		}

		if err := writer.WriteNewLine(); err != nil {
			return err
		}
	}

	return writer.Flush()
}
