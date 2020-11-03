package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type ParseResult struct {
	Ranges []Range
}

// Select select columns
func (pr ParseResult) Select(s []string) ([]string, error) {
	var rt []string
	l := len(s)
	for _, r := range pr.Ranges {
		for idx := range r.Enumerate(l) {
			if idx == 0 {
				// index zero is all columns. (like awk)
				rt = append(rt, s...)
			} else if 0 <= idx-1 && idx-1 < l {
				// select index
				rt = append(rt, s[idx-1])
			} else {
				return nil, fmt.Errorf("index out of range")
			}
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
	var ranges []Range
	for _, v := range p.query {
		r, err := newRange(v)
		if err != nil {
			return ParseResult{}, err
		}

		ranges = append(ranges, r)
	}

	return ParseResult{Ranges: ranges}, nil
}

type Range struct {
	start int
	stop  index
	step  int
}

type index struct {
	num int
	inf bool
}

func newPositionIndex(idx int) index {
	return index{num: idx, inf: false}
}

func newInfIndex() index {
	return index{inf: true}
}

func newRange(query string) (Range, error) {
	var err error
	split := strings.Split(query, ":")
	if len(split) == 1 {
		// \d+
		idx, err := strconv.Atoi(split[0])
		return Range{
			start: idx,
			stop:  newPositionIndex(idx),
			step:  1,
		}, err
	} else if len(split) == 2 || len(split) == 3 {
		// \d*:\d*:\d*
		start := 1
		if len(split[0]) != 0 {
			idx, err := strconv.Atoi(split[0])
			if err != nil {
				return Range{}, err
			}
			start = idx
		}

		stop := newInfIndex()
		if len(split[1]) != 0 {
			idx, err := strconv.Atoi(split[1])
			if err != nil {
				return Range{}, err
			}
			stop = newPositionIndex(idx)
		}

		step := 1
		if len(split) == 3 && len(split[2]) != 0 {
			step, err = strconv.Atoi(split[2])
			if err != nil {
				return Range{}, err
			}
		}

		if step == 0 {
			return Range{}, fmt.Errorf("step cannot be zero")
		}

		return Range{start: start, stop: stop, step: step}, nil
	}

	return Range{}, fmt.Errorf("failed to parse query: %s", query)
}

// Enumerate enumerate index
func (r Range) Enumerate(max int) <-chan int {
	rt := make(chan int)

	go func() {
		defer close(rt)

		start := r.start
		if start < 0 {
			start = max + start + 1
		}

		stop := r.stop.num
		if r.stop.inf || stop >= max {
			stop = max
		}
		if stop < 0 {
			stop = max + stop + 1
		}

		step := r.step

		if start == stop {
			rt <- start
		} else if start < stop {
			if step < 0 {
				return
			}
			for idx := start; idx <= stop; idx += step {
				rt <- idx
			}
		} else {
			if step > 0 {
				return
			}
			for idx := start; idx >= stop; idx += step {
				rt <- idx
			}
		}

	}()

	return rt
}
