package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xztaityozx/sel/internal/column"
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/option"
	"github.com/xztaityozx/sel/internal/output"
	"github.com/xztaityozx/sel/internal/parser"
)

var Version string = "undefined"

// stdinSourceName は標準入力から読んでいるときにエラーメッセージで使うソース名
const stdinSourceName = "<stdin>"

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
	Args:          cobra.MinimumNArgs(1),
	Version:       Version,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// ここから先は実行時エラーなので、Usage は出さない（引数のパースエラーはこの手前で弾かれる）
		cmd.SilenceUsage = true

		opt, err := option.NewOption(viper.GetViper())
		if err != nil {
			return err
		}
		selectors, err := parser.Parse(args)
		if err != nil {
			return err
		}

		w := output.NewWriter(opt, os.Stdout, false)

		if len(opt.Files) != 0 {
			files, err := opt.Enumerate()
			if err != nil {
				return err
			}

			for _, file := range files {
				fp, err := os.Open(file)
				if err != nil {
					return err
				}
				if err := run(fp, file, opt, w, selectors, args); err != nil {
					return err
				}
			}

			return nil
		}

		return run(os.Stdin, stdinSourceName, opt, w, selectors, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "sel:", err)
		os.Exit(1)
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
	rootCmd.Flags().BoolP(option.NameIgnoreMissing, "M", false, "output empty string for out-of-range columns instead of error")
	rootCmd.Flags().StringP(option.NameFillMissing, "E", option.DefaultFillMissing, "fill value for out-of-range columns (implies -M)")
	rootCmd.Flags().Bool(option.NameCsv, false, "parse input file as CSV")
	rootCmd.Flags().Bool(option.NameTsv, false, "parse input file as TSV")
	rootCmd.Flags().StringP(option.NameTemplate, "t", option.DefaultTemplate, "template for output")
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
		"$ cat /path/to/file.csv | sel --csv 1 2 3 4",
		"$ sel 2:: -f ./file",
		"$ cat /path/to/file | sel /^begin/:/^end/",
		"$ echo AAA BBB CCC | sel --template 'one: {} two: {} three: {}' 1 2 3",
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

// positionError はエラーに発生位置（入力元・行番号・クエリ）を付与する
type positionError struct {
	source string
	line   int
	query  string
	err    error
}

func (e *positionError) Error() string {
	if e.query != "" {
		return fmt.Sprintf("%s:%d: query %q: %s", e.source, e.line, e.query, e.err)
	}
	return fmt.Sprintf("%s:%d: %s", e.source, e.line, e.err)
}

func (e *positionError) Unwrap() error { return e.err }

// run はあるファイルについて column.Selector によるカラム選択と column.Writer による書き出しを行う。ファイルはCloseされる
// source はエラーメッセージに出す入力元の名前（ファイルパスか stdinSourceName）
// queries は selectors と同じ順番のクエリ文字列で、エラーの発生位置を表すのに使う
func run(input *os.File, source string, opt option.Option, w *output.Writer, selectors []column.Selector, queries []string) error {
	defer func(input *os.File) {
		_ = input.Close()
	}(input)

	src, err := iterator.NewSource(opt, input)
	if err != nil {
		return err
	}

	var filler missingFiller
	if opt.IgnoreMissing {
		filler = missingFiller{enabled: true, fill: []byte(opt.FillMissing)}
	}

	line := 0
	for {
		columns, err := src.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return &positionError{source: source, line: line + 1, err: err}
		}
		line++

		if err := selectAll(columns, w, selectors, queries, filler); err != nil {
			return &positionError{source: source, line: line, query: err.query, err: err.err}
		}
	}

	if err := w.Flush(); err != nil {
		return &positionError{source: source, line: line, err: err}
	}

	return nil
}

// selectError は selectAll 内でどのクエリが失敗したかを保持する
type selectError struct {
	query string
	err   error
}

func (e *selectError) Error() string { return e.err.Error() }
func (e *selectError) Unwrap() error { return e.err }

// missingFiller は範囲外のカラムをどう埋めるかを表す。
// enabled が false のときは範囲外をエラーにする(-M も -E も指定されていない)
type missingFiller struct {
	enabled bool
	// 埋める値。長さ0のときは何も書き出さない。
	// 行ごとに []byte へ変換しなおさなくていいように、run() で一度だけ作る
	fill []byte
}

func selectAll(columns iterator.Columns, w *output.Writer, selectors []column.Selector, queries []string, filler missingFiller) *selectError {
	for i, selector := range selectors {
		err := selector.Select(w, columns)
		if err != nil {
			if filler.enabled && iterator.IsIndexOutOfRange(err) {
				if len(filler.fill) != 0 {
					if werr := w.Write(filler.fill); werr != nil {
						return &selectError{query: queries[i], err: werr}
					}
				}
				continue
			}
			if iterator.IsIndexOutOfRange(err) {
				err = fmt.Errorf("line has only %d columns", len(columns.ToArray()))
			}
			return &selectError{query: queries[i], err: err}
		}
	}
	if err := w.WriteNewLine(); err != nil {
		return &selectError{err: err}
	}
	return nil
}
