package parser

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	as := assert.New(t)

	t.Run("引数が0でもいい", func(t *testing.T) {
		expect := Parser{query: nil}
		actual := New()

		as.Equal(expect, actual)
	})

	t.Run("引数が1でもいい", func(t *testing.T) {
		expect := Parser{query: []string{"1"}}
		actual := New("1")

		as.Equal(expect, actual)
	})

	t.Run("引数が何個でもいい", func(t *testing.T) {
		expect := Parser{query: []string{"1", "2", "3"}}
		actual := New("1", "2", "3")

		as.Equal(expect, actual)
	})
}

func TestParser_Parse(t *testing.T) {
	as := assert.New(t)

	data := []struct {
		in     string
		expect []int
	}{
		{in: "1", expect: []int{1}},
		{in: "1 2 3", expect: []int{1, 2, 3}},
		{in: "2 3 1", expect: []int{2, 3, 1}},
		{in: "1:3", expect: []int{1, 2, 3}},
		{in: "1 2:4", expect: []int{1, 2, 3, 4}},
		{in: "1:2 3:4", expect: []int{1, 2, 3, 4}},
		{in: "1:3 4", expect: []int{1, 2, 3, 4}},
		{in: "4:1", expect: []int{4, 3, 2, 1}},
		{in: "4:2 1", expect: []int{4, 3, 2, 1}},
		{in: "1:1:4", expect: []int{1, 2, 3, 4}},
		{in: "4:1:1", expect: []int{4, 3, 2, 1}},
		{in: "4:1:1 1:1:4", expect: []int{4, 3, 2, 1, 1, 2, 3, 4}},
		{in: "1:1:4 4:1:1", expect: []int{1, 2, 3, 4, 4, 3, 2, 1}},
		{in: "1:2:10", expect: []int{1, 3, 5, 7, 9}},
		{in: "1:3:10", expect: []int{1, 4, 7, 10}},
		{in: "1:2:10 1", expect: []int{1, 3, 5, 7, 9, 1}},
	}

	for _, v := range data {
		split := strings.Split(v.in, " ")
		pr, err := New(split...).Parse()

		as.Nil(err)
		as.Equal(v.expect, pr.SelectedColumns, fmt.Sprintf("query: %v", v.in))
	}
}

func TestParseResult_Select(t *testing.T) {
	as := assert.New(t)

	line := []string{"item1", "item2", "item3"}

	t.Run("indexに負数があると", func(t *testing.T) {
		_, err := ParseResult{SelectedColumns: []int{1, 2, -1}}.Select(line)
		as.Error(err, "例外が投げられる")
	})

	t.Run("indexに範囲を超えるものがあると", func(t *testing.T) {
		_, err := ParseResult{SelectedColumns: []int{1, 2, len(line) + 1}}.Select(line)
		as.Error(err, "例外が投げられる")
	})

	t.Run("indexに0があると", func(t *testing.T) {
		actual, err := ParseResult{SelectedColumns: []int{1, 2, 0}}.Select(line)
		as.Nil(err)
		as.Equal(append([]string{line[0], line[1]}, line...), actual, "全部が選ばれる")
	})

	t.Run("同じ数字を選んでもよい", func(t *testing.T) {
		actual, err := ParseResult{SelectedColumns: []int{1, 1, 1}}.Select(line)
		as.Nil(err)
		as.Equal([]string{line[0], line[0], line[0]}, actual)
	})
}
