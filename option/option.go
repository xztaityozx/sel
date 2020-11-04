package option

import (
	"fmt"
	"os"
	"path/filepath"
)

// Option is commandline options
type Option struct {
	// -i, --in-place option
	InPlace bool
	// -b, --backup
	Backup bool
	// --input/output-delimiter option
	DelimiterOption
	// -f, --input-files option
	InputFiles
}

// DelimiterOption is setting for --input/output-delimiter option
type DelimiterOption struct {
	// --input-delimiter
	Input string
	// --output-delimiter
	OutPut string
	// --remove-empty
	RemoveEmpty bool
}

// InputFiles is setting for -f, --input-files option
type InputFiles struct {
	Files []string
}

// Enumerate enumerate /path/to/input/files
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
