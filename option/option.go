package option

import (
	"fmt"
	"os"
	"path/filepath"

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
func NewOption(v *viper.Viper) Option {

	if v.GetBool(NameFieldSplit) {
		return Option{
			DelimiterOption: DelimiterOption{
				InputDelimiter:  `\s+`,
				OutPutDelimiter: v.GetString(NameOutPutDelimiter),
				RemoveEmpty:     v.GetBool(NameRemoveEmpty),
				UseRegexp:       true,
				SplitBefore:     v.GetBool(NameSplitBefore),
			},
			InputFiles: InputFiles{v.GetStringSlice(NameInputFiles)},
			Xsv: Xsv{
				Csv: v.GetBool(NameCsv),
				Tsv: v.GetBool(NameTsv),
			},
		}
	}

	return Option{
		DelimiterOption: DelimiterOption{
			InputDelimiter:  v.GetString(NameInputDelimiter),
			OutPutDelimiter: v.GetString(NameOutPutDelimiter),
			RemoveEmpty:     v.GetBool(NameRemoveEmpty),
			UseRegexp:       v.GetBool(NameUseRegexp),
			SplitBefore:     v.GetBool(NameSplitBefore),
		},
		InputFiles: InputFiles{v.GetStringSlice(NameInputFiles)},
		Xsv: Xsv{
			Csv: v.GetBool(NameCsv),
			Tsv: v.GetBool(NameTsv),
		},
	}
}
