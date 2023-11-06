package option

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"
)

// Option is commandline options
type Option struct {
	// --input/output-delimiter option
	DelimiterOption
	// -f, --input-files option
	InputFiles
	// XSV support
	Xsv
	// --template
	Template *template.Template
}

// DelimiterOption is setting for --input/output-delimiter option
type DelimiterOption struct {
	// --input-delimiter
	InputDelimiter string
	// --output-delimiter
	OutPutDelimiter string
	// --remove-empty
	RemoveEmpty bool
	// --use-regexp
	UseRegexp bool
	// --split-before
	SplitBefore bool
}

// InputFiles is setting for -f, --input-files option
type InputFiles struct {
	Files []string
}

// Enumerate /path/to/input/files
func (ifs InputFiles) Enumerate() ([]string, error) {
	if ifs.Files == nil || len(ifs.Files) == 0 {
		return nil, fmt.Errorf("there are no files")
	}

	var rt []string

	for _, v := range ifs.Files {
		expanded, err := filepath.Glob(v)
		if err != nil {
			return nil, err
		}

		for _, p := range expanded {
			fi, err := os.Stat(p)
			if err != nil {
				return nil, err
			} else if !fi.Mode().IsRegular() {
				return nil, fmt.Errorf("%s is not regular file", p)
			} else if fi.IsDir() {
				return nil, fmt.Errorf("%s is directory", p)
			}

			rt = append(rt, p)
		}
	}

	if len(rt) == 0 {
		return nil, fmt.Errorf("no files(path/glob is wrong?)")
	}

	return rt, nil
}

const (
	NameInputDelimiter  = "input-delimiter"
	NameOutPutDelimiter = "output-delimiter"
	NameRemoveEmpty     = "remove-empty"
	NameUseRegexp       = "use-regexp"
	NameInputFiles      = "input-files"
	NameSplitBefore     = "split-before"
	NameFieldSplit      = "field-split"
	NameCsv             = "csv"
	NameTsv             = "tsv"
	NameTemplate        = "template"

	DefaultTemplate = ""
)

type SplitStrategy int

func GetOptionNames() []string {
	return []string{
		NameInputFiles,
		NameInputDelimiter,
		NameOutPutDelimiter,
		NameUseRegexp,
		NameRemoveEmpty,
		NameSplitBefore,
		NameFieldSplit,
		NameCsv,
		NameTsv,
		NameTemplate,
	}
}

// Xsv is option group for xsv support
type Xsv struct {
	Csv bool
	Tsv bool
}

func (x Xsv) IsXsv() (bool, rune) {
	if x.Csv {
		return true, ','
	} else if x.Tsv {
		return true, '\t'
	} else {
		return false, ','
	}
}

// NewOption は viper.Viper からフラグの値を取り出して Option を作って返す
func NewOption(v *viper.Viper) (Option, error) {

	// 入力のデリミターの受け取り、NameFieldSplit が指定されているときは `\s+` で上書き
	inputDelimiter := v.GetString(NameInputDelimiter)
	if v.GetBool(NameFieldSplit) {
		inputDelimiter = `\s+`
	}

	// デリミターを正規表現として処理するかどうか。NameFieldSplit が指定されているときは強制的にON
	useRegexp := v.GetBool(NameUseRegexp) || v.GetBool(NameFieldSplit)

	// --templateオプションで出力のフォーマットを指定するやつ
	var tmpl *template.Template
	if v.GetString(NameTemplate) != DefaultTemplate {
		var err error
		tmpl, err = parseTemplate(v.GetString(NameTemplate))
		if err != nil {
			return Option{}, err
		}
	}

	return Option{
		DelimiterOption: DelimiterOption{
			InputDelimiter:  inputDelimiter,
			OutPutDelimiter: v.GetString(NameOutPutDelimiter),
			RemoveEmpty:     v.GetBool(NameRemoveEmpty),
			UseRegexp:       useRegexp,
			SplitBefore:     v.GetBool(NameSplitBefore),
		},
		InputFiles: InputFiles{v.GetStringSlice(NameInputFiles)},
		Xsv: Xsv{
			Csv: v.GetBool(NameCsv),
			Tsv: v.GetBool(NameTsv),
		},
		Template: tmpl,
	}, nil
}

func parseTemplate(input string) (*template.Template, error) {
	var result string

	// input ::= char | marker | input
	// char ::= 任意の文字
	// marker ::= {}
	cnt := 0
	for i := 0; i < len(input); i++ {
		switch {
		case input[i] == '{' && i+1 < len(input) && input[i+1] == '}':
			result += fmt.Sprintf("{{ index . %d }}", cnt)
			cnt++
			i++
		default:
			result += string(input[i])
		}
	}

	return template.New("output").Parse(result)
}
