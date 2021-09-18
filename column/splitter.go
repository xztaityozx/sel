package column

import (
	"regexp"
	"strings"
)

type Splitter struct {
	reg         *regexp.Regexp
	str         string
	removeEmpty bool
}

func NewSplitter(str string) Splitter {
	return Splitter{reg: nil, str: str}
}

func NewSplitterRegexp(query string) (Splitter, error) {
	r, e := regexp.Compile(query)
	return Splitter{reg: r}, e
}

func (s Splitter) Split(line string) []string {
	if s.reg == nil {
		return s.removeEmptyColumn(strings.Split(line, s.str))
	} else {
		return s.removeEmptyColumn(s.reg.Split(line, -1))
	}
}

func (s Splitter) removeEmptyColumn(in []string) []string {
	if !s.removeEmpty {
		return in
	}

	rt := make([]string, 0, len(in))
	for _, v := range in {
		if len(v) != 0 {
			rt = append(rt, v)
		}
	}

	return rt
}
