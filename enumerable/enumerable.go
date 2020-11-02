package enumerable

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Range struct {
	start index
	stop  index
	step  int
}

type index struct {
	num int
	inf bool
}

func NewPositionIndex(idx int) index {
	return index{num: idx, inf: false}
}

func NewInfIndex() index {
	return index{inf: true}
}

var dic = []struct {
	re      *regexp.Regexp
	builder func([]int) Range
}{
	{
		re: regexp.MustCompile(`^\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewPositionIndex(s[0]),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+:\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewPositionIndex(s[1]),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+:\d+:\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewPositionIndex(s[1]),
				step:  s[2],
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+:$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewInfIndex(),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^:\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewPositionIndex(s[1]),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^:$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewInfIndex(),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+:\d+:$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewPositionIndex(s[1]),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+::$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewInfIndex(),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^::$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewInfIndex(),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^:\d+:$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewPositionIndex(s[1]),
				step:  1,
			}
		},
	},
	{
		re: regexp.MustCompile(`^:\d+:\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewPositionIndex(s[1]),
				step:  s[2],
			}
		},
	},
	{
		re: regexp.MustCompile(`^::\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(1),
				stop:  NewInfIndex(),
				step:  s[2],
			}
		},
	},
	{
		re: regexp.MustCompile(`^\d+::\d+$`),
		builder: func(s []int) Range {
			return Range{
				start: NewPositionIndex(s[0]),
				stop:  NewInfIndex(),
				step:  s[2],
			}
		},
	},
}

func NewRange(query string) (Range, error) {

	var split []int
	for _, v := range strings.Split(query, ":") {
		if len(v) == 0 {
			split = append(split, 0)
			continue
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return Range{}, err
		}
		split = append(split, i)
	}

	for _, v := range dic {
		if v.re.MatchString(query) {
			r := v.builder(split)
			if r.step == 0 {
				return Range{}, fmt.Errorf("step is zero")
			}

			return r, nil
		}
	}

	return Range{}, fmt.Errorf("failed to parse query: %s", query)
}
