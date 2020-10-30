package enumerable

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Range struct {
	start       position
	stop        position
	step        position
	ExpectedNil bool
}

type position struct {
	num int
	inf bool
}

func NewPositionIndex(idx int) position {
	return position{num: idx, inf: false}
}

func NewInf() position {
	return position{inf: true}
}

var d = regexp.MustCompile(`^\d+$`)
var dd = regexp.MustCompile(`^\d+:\d+$`)
var ddd = regexp.MustCompile(`^\d+:\d+:\d+$`)
var d_ = regexp.MustCompile(`^\d+:$`)
var _d = regexp.MustCompile(`^:\d+$`)
var _dd = regexp.MustCompile(`^:\d+:\d+$`)
var __d = regexp.MustCompile(`^::\d+$`)
var _d_ = regexp.MustCompile(`^:\d+:$`)
var dd_ = regexp.MustCompile(`^\d+:\d+:$`)
var d__ = regexp.MustCompile(`^\d+::$`)

func NewRange(query string) (Range, error) {

	var split []int
	for _, v := range strings.Split(query, ":") {
		i, err := strconv.Atoi(v)
		if err != nil {
			return Range{}, err
		}
		split = append(split, i)
	}

	return Range{}, fmt.Errorf("failed to parse query: %s", query)

}
