package paser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ParseResult struct {
	SelectedColumns []int
}

// Select select columns
func (pr ParseResult) Select(s []string) ([]string, error) {
	var rt []string
	l := len(s)
	for _, idx := range pr.SelectedColumns {
		if l < idx || idx < 0 {
			return nil, fmt.Errorf("index out of range: selected=%d, length=%d", idx, l)
		} else if idx == 0 {
			// index 0 is line. like awk
			rt = append(rt, s...)
		} else {
			rt = append(rt, s[idx-1])
		}
	}

	return rt, nil
}

type Parser struct {
	query []string
}

func New(q ...string) Parser {
	return Parser{query: q}
}

func (p Parser) Parse() (ParseResult, error) {
	var enumerated []int

	c := regexp.MustCompile(`^\d+$`)
	se := regexp.MustCompile(`^\d+:\d+$`)
	sse := regexp.MustCompile(`^\d+:\d+:\d+$`)

	var start, step, stop int

	for _, v := range p.query {
		split := strings.Split(v, ":")

		if c.MatchString(v) {
			// \d+ column
			// Ex) 1
			start, _ = strconv.Atoi(split[0])
			step = 1
			stop = start
		} else if se.MatchString(v) {
			// \d+:\d+ start:stop
			// Ex) 1:2
			start, _ = strconv.Atoi(split[0])
			step = 1
			stop, _ = strconv.Atoi(split[1])
		} else if sse.MatchString(v) {
			// \d+:\d+:\d+ start:step:stop
			// Ex) 1:2:3
			start, _ = strconv.Atoi(split[0])
			step, _ = strconv.Atoi(split[1])
			stop, _ = strconv.Atoi(split[2])
		} else {
			return ParseResult{}, fmt.Errorf("%s is invalid query", v)
		}

		e, err := enumerate(start, step, stop)
		if err != nil {
			return ParseResult{}, err
		}
		enumerated = append(enumerated, e...)
	}

	return ParseResult{SelectedColumns: enumerated}, nil
}

func enumerate(start, step, stop int) ([]int, error) {
	if step <= 0 {
		return nil, fmt.Errorf("step must be greater than zero")
	}

	if start < 0 || stop < 0 {
		return nil, fmt.Errorf("column number must be greater equal zero")
	}

	var rt []int
	if stop < start {
		for i := start; i >= stop; i -= step {
			rt = append(rt, i)
		}
	} else if start < stop {
		for i := start; i <= stop; i += step {
			rt = append(rt, i)
		}
	} else {
		rt = append(rt, start)
	}

	return rt, nil
}
