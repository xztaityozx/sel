package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert.Equal(t, Parser{query: []string{"query"}}, New("query"))
}

func TestParser_Parse(t *testing.T) {
	max := 10
	as := assert.New(t)

	t.Run("正常系", func(t *testing.T) {
		data := map[string][]Range{
			"1": {Range{start: 1, stop: newPositionIndex(1), step: 1}},
			"1:10 :10 1: : ::": {
				Range{start: 1, stop: newPositionIndex(max), step: 1},
				Range{start: 1, stop: newPositionIndex(max), step: 1},
				Range{start: 1, stop: newInfIndex(), step: 1},
				Range{start: 1, stop: newInfIndex(), step: 1},
				Range{start: 1, stop: newInfIndex(), step: 1},
			},
			"1:5:1 1:10:2": {
				Range{start: 1, stop: newPositionIndex(5), step: 1},
				Range{start: 1, stop: newPositionIndex(max), step: 2},
			},
		}

		for q, expect := range data {
			pr, err := New(strings.Split(q, " ")...).Parse()
			as.Nil(err)
			as.Equalf(expect, pr.Ranges, "%v", q)
		}
	})

	t.Run("異常系", func(t *testing.T) {
		for _, q := range []string{
			"1:10:0",
			"invalid",
			"invalid:1",
			"1:invalid",
			"1:1:invalid",
		} {
			_, err := New(q).Parse()
			as.Error(err)
		}
	})
}

func TestParseResult_Select(t *testing.T) {
	as := assert.New(t)

	line := func() []string {
		var rt []string
		for i := 0; i < 10; i++ {
			rt = append(rt, fmt.Sprint(i+1))
		}
		return rt
	}()

	t.Run("正常系", func(t *testing.T) {
		data := map[string][]int{
			"1":                  {1},
			"1 2 3":              {1, 2, 3},
			"1:5":                {1, 2, 3, 4, 5},
			":":                  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			"::":                 {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			"1:4 4:1:-1":         {1, 2, 3, 4, 4, 3, 2, 1},
			"0":                  {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			"1 1":                {1, 1},
			"-4: 0 :-4":          {7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, 4, 5, 6, 7},
			"1 5:1 5:1:1 5:1:-1": {1, 5, 4, 3, 2, 1},
			"11":                 {},
		}

		for q, d := range data {
			pr, err := New(strings.Split(q, " ")...).Parse()
			as.Nil(err)
			expect := func() []string {
				var rt []string
				for _, v := range d {
					rt = append(rt, fmt.Sprint(v))
				}
				return rt
			}()

			actual, err := pr.Select(line)
			as.Nil(err)

			as.Equalf(expect, actual, "%s", q)
		}
	})
}

func TestNewRange(t *testing.T) {
	as := assert.New(t)
	data := map[string]Range{
		"1":      {start: 1, stop: newPositionIndex(1), step: 1},
		"1:10":   {start: 1, stop: newPositionIndex(10), step: 1},
		"1:":     {start: 1, stop: newInfIndex(), step: 1},
		":":      {start: 1, stop: newInfIndex(), step: 1},
		":4":     {start: 1, stop: newPositionIndex(4), step: 1},
		"2:10:3": {start: 2, stop: newPositionIndex(10), step: 3},
		"3:10:":  {start: 3, stop: newPositionIndex(10), step: 1},
		"4::10":  {start: 4, stop: newInfIndex(), step: 10},
		"5::":    {start: 5, stop: newInfIndex(), step: 1},
		"::":     {start: 1, stop: newInfIndex(), step: 1},
		":6:":    {start: 1, stop: newPositionIndex(6), step: 1},
		":6:3":   {start: 1, stop: newPositionIndex(6), step: 3},
		"::3":    {start: 1, stop: newInfIndex(), step: 3},
		"-6":     {start: -6, stop: newPositionIndex(-6), step: 1},
		"-3:":    {start: -3, stop: newInfIndex(), step: 1},
	}

	for q, expect := range data {
		actual, err := newRange(q)
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
		{q: "1::", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: ":10:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: ":10:", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "::", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "::1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: ":", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "::3", max: 10, expect: []int{1, 4, 7, 10}},
		{q: "10:", max: 10, expect: []int{10}},
		{q: "1:20", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{q: "-4::1", max: 10, expect: []int{7, 8, 9, 10}},
		{q: ":-4:1", max: 10, expect: []int{1, 2, 3, 4, 5, 6, 7}},
		{q: "-8:-4:1", max: 10, expect: []int{3, 4, 5, 6, 7}},
		{q: "-4:-8:-1", max: 10, expect: []int{7, 6, 5, 4, 3}},
		{q: "1:2:3", max: 10, expect: []int{1}},
		{q: "1:1", max: 10, expect: []int{1}},
		{q: "1:1:1", max: 10, expect: []int{1}},
		{q: "1:4:-1", max: 10, expect: nil},
		{q: "4:3:1", max: 10, expect: nil},
		{q: "4:3:", max: 10, expect: nil},
		{q: "4:3", max: 10, expect: nil},
		{q: "::-1", max: 10, expect: nil},
		{q: "8:4:1", max: 10, expect: nil},
		{q: "-8:-4:-1", max: 10, expect: nil},
	}

	for _, v := range data {
		r, err := newRange(v.q)
		as.Nil(err)
		var actual []int
		for _, idx := range r.Enumerate(v.max) {
			actual = append(actual, idx)
		}

		as.Equalf(v.expect, actual, "%s", v.q)
	}
}
