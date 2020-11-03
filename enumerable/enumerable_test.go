package enumerable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRange(t *testing.T) {
	as := assert.New(t)
	data := map[string]Range{
		"1":      {start: 1, stop: NewPositionIndex(1), step: 1},
		"1:10":   {start: 1, stop: NewPositionIndex(10), step: 1},
		"1:":     {start: 1, stop: NewInfIndex(), step: 1},
		":":      {start: 1, stop: NewInfIndex(), step: 1},
		":4":     {start: 1, stop: NewPositionIndex(4), step: 1},
		"2:10:3": {start: 2, stop: NewPositionIndex(10), step: 3},
		"3:10:":  {start: 3, stop: NewPositionIndex(10), step: 1},
		"4::10":  {start: 4, stop: NewInfIndex(), step: 10},
		"5::":    {start: 5, stop: NewInfIndex(), step: 1},
		"::":     {start: 1, stop: NewInfIndex(), step: 1},
		":6:":    {start: 1, stop: NewPositionIndex(6), step: 1},
		":6:3":   {start: 1, stop: NewPositionIndex(6), step: 3},
		"::3":    {start: 1, stop: NewInfIndex(), step: 3},
		"-6":     {start: -6, stop: NewPositionIndex(-6), step: 1},
		"-3:":    {start: -3, stop: NewInfIndex(), step: 1},
	}

	for q, expect := range data {
		actual, err := NewRange(q)
		as.Nil(err)
		as.Equal(expect, actual, fmt.Sprintf("query: %s", q))
	}

}

func TestRange_Enumerate(t *testing.T) {
	as := assert.New(t)

	data := []struct {
		q      string
		max    int
		expect []int
	}{
		{q: "1:10", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "1", max: 10, expect: []int{1}},
		{q: "10", max: 10, expect: []int{10}},
		{q: "1:", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: ":10", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "3:7", max: 10, expect: []int{3, 4, 5, 6, 7}},
		{q: "-1", max: 10, expect: []int{10}},
		{q: "-3:", max: 10, expect: []int{8, 9, 10}},
		{q: ":-3", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8}},
		{q: "1:10:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "1:10:", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "1:10:3", max: 10, expect: []int{1, 4, 7, 10}},
		{q: "10:1:-1", max: 10, expect: []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{q: "8:4:-1", max: 10, expect: []int{8, 7, 6, 5, 4}},
		{q: "1:10:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "1:10:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "1:10:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, v := range data {
		r, err := NewRange(v.q)
		as.Nil(err)
		var actual []int
		for idx := range r.Enumerate(v.max) {
			actual = append(actual, idx)
		}

		as.Equalf(v.expect, actual, "%s", v.q)
	}
}
