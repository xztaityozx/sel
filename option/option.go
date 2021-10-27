package option

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Option is commandline options
type Option struct {
	// --input/output-delimiter option
	DelimiterOption
	// -f, --input-files option
	InputFiles
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
	NameInputFiles  = "input-files"
	NameSplitBefore = "split-before"
)

type SplitStrategy int

const (
	SplitStrategyPreSplit SplitStrategy = iota
	SplitStrategyLazySplit
)

func GetOptionNames() []string {
	return []string{
		NameInputFiles,
		NameInputDelimiter,
		NameOutPutDelimiter,
		NameUseRegexp,
		NameRemoveEmpty,
		NameSplitBefore,
	}
}

// NewOption は viper.Viper からフラグの値を取り出して Option を作って返す
func NewOption(v *viper.Viper) Option {
	return Option{
		DelimiterOption: DelimiterOption{
			InputDelimiter:  v.GetString(NameInputDelimiter),
			OutPutDelimiter: v.GetString(NameOutPutDelimiter),
			RemoveEmpty:     v.GetBool(NameRemoveEmpty),
			UseRegexp:       v.GetBool(NameUseRegexp),
			SplitBefore:     v.GetBool(NameSplitBefore),
		},
		InputFiles: InputFiles{v.GetStringSlice(NameInputFiles)},
	}
}
