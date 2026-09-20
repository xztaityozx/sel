package column

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xztaityozx/sel/internal/iterator"
)

// apply は cols に対して -x 相当の除外をかけ、残ったカラムを文字列で返す
func apply(t *testing.T, cols []string, selectors ...Selector) ([]string, error) {
	t.Helper()

	e, err := NewExclusion(selectors, nil)
	require.NoError(t, err)

	columns, err := e.Apply(&testColumns{a: cols})
	if err != nil {
		return nil, err
	}

	var rt []string
	for _, v := range columns.ToArray() {
		rt = append(rt, string(v))
	}
	return rt, nil
}

func TestExclusion_Apply(t *testing.T) {
	cols := []string{"a", "b", "c", "d", "e"}

	tests := []struct {
		name      string
		selectors []Selector
		want      []string
	}{
		{name: "index", selectors: []Selector{NewIndexSelector(2)}, want: []string{"a", "c", "d", "e"}},
		{name: "負のindex", selectors: []Selector{NewIndexSelector(-1)}, want: []string{"a", "b", "c", "d"}},
		{name: "範囲外のindexは無視", selectors: []Selector{NewIndexSelector(9)}, want: cols},
		{name: "複数", selectors: []Selector{NewIndexSelector(2), NewIndexSelector(4)}, want: []string{"a", "c", "e"}},
		{name: "重複しても1回", selectors: []Selector{NewIndexSelector(2), NewIndexSelector(2)}, want: []string{"a", "c", "d", "e"}},
		{name: "range", selectors: []Selector{NewRangeSelector(2, 1, 4, false)}, want: []string{"a", "e"}},
		{name: "開いたrange", selectors: []Selector{NewRangeSelector(2, 1, 2, true)}, want: []string{"a"}},
		{name: "step", selectors: []Selector{NewRangeSelector(1, 2, 5, false)}, want: []string{"b", "d"}},
		{name: "逆順range", selectors: []Selector{NewRangeSelector(4, -1, 2, false)}, want: []string{"a", "e"}},
		{name: "負のrange", selectors: []Selector{NewRangeSelector(-2, 1, -1, false)}, want: []string{"a", "b", "c"}},
		{name: "行より後ろのrangeは無視", selectors: []Selector{NewRangeSelector(10, 1, 12, false)}, want: cols},
		{name: "行より後ろの開いたrangeは無視", selectors: []Selector{NewRangeSelector(10, 1, 10, true)}, want: cols},
		{name: "全部除外", selectors: []Selector{NewRangeSelector(1, 1, 1, true)}, want: nil},
		// 行より後ろまで伸びた範囲は、行の外を舐めずに行内だけ落とす(舐めると行ごとに巨大なループになる)
		{name: "行をはみ出すrange", selectors: []Selector{NewRangeSelector(2, 1, 100000000000, false)}, want: []string{"a"}},
		{name: "行をはみ出すrangeとstep", selectors: []Selector{NewRangeSelector(1, 3, math.MaxInt64, false)}, want: []string{"b", "c", "e"}},
		{name: "行をはみ出す逆順range", selectors: []Selector{NewRangeSelector(5, -1, -1000000000000, false)}, want: nil},
		{name: "行をはみ出す逆順rangeとstep", selectors: []Selector{NewRangeSelector(math.MaxInt64, -2, 1, false)}, want: []string{"b", "d"}},
		// start == stop は1カラムを指すだけなので step の向きは問わない(Select と同じ)
		{name: "start == stop で負のstep", selectors: []Selector{NewRangeSelector(2, -1, 2, false)}, want: []string{"a", "c", "d", "e"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apply(t, cols, tt.selectors...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("空行", func(t *testing.T) {
		got, err := apply(t, nil, NewIndexSelector(1), NewRangeSelector(2, 1, 3, false), NewRangeSelector(2, 1, 2, true))
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("stepの向きが違うとエラー", func(t *testing.T) {
		_, err := apply(t, cols, NewRangeSelector(1, -1, 5, false))
		require.Error(t, err)

		_, err = apply(t, cols, NewRangeSelector(5, 1, 1, false))
		require.Error(t, err)
	})

	t.Run("除外後のカラムは1から番号付けし直される", func(t *testing.T) {
		e, err := NewExclusion([]Selector{NewIndexSelector(2)}, nil)
		require.NoError(t, err)

		columns, err := e.Apply(&testColumns{a: cols})
		require.NoError(t, err)

		first, err := columns.ElementAt(1)
		require.NoError(t, err)
		assert.Equal(t, "a", string(first))

		second, err := columns.ElementAt(2)
		require.NoError(t, err)
		assert.Equal(t, "c", string(second))

		last, err := columns.ElementAt(-1)
		require.NoError(t, err)
		assert.Equal(t, "e", string(last))

		_, err = columns.ElementAt(5)
		assert.True(t, iterator.IsIndexOutOfRange(err))
	})

	t.Run("行をまたいでバッファを使い回す", func(t *testing.T) {
		e, err := NewExclusion([]Selector{NewIndexSelector(1)}, nil)
		require.NoError(t, err)

		for _, line := range [][]string{{"a", "b", "c"}, {"d"}, nil, {"e", "f"}} {
			columns, err := e.Apply(&testColumns{a: line})
			require.NoError(t, err)

			var got []string
			for _, v := range columns.ToArray() {
				got = append(got, string(v))
			}
			if len(line) <= 1 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, line[1:], got, "line: %s", strings.Join(line, " "))
			}
		}
	})
}

func TestNewExclusion(t *testing.T) {
	t.Run("switchクエリは除外に使えない", func(t *testing.T) {
		sw, err := NewSwitchSelector("/a/", "/b/")
		require.NoError(t, err)

		_, err = NewExclusion([]Selector{sw}, []string{"/a/:/b/"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "/a/:/b/")
	})

	t.Run("index 0 は除外に使えない", func(t *testing.T) {
		_, err := NewExclusion([]Selector{NewIndexSelector(0)}, []string{"0"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"0"`)
	})

	t.Run("rangeに埋まったindex 0 も除外に使えない", func(t *testing.T) {
		for _, tt := range []struct {
			query string
			sel   RangeSelector
		}{
			{query: "0:0", sel: NewRangeSelector(0, 1, 0, false)},
			{query: "0:2", sel: NewRangeSelector(0, 1, 2, false)},
			{query: "2:0", sel: NewRangeSelector(2, -1, 0, false)},
			{query: "0:", sel: NewRangeSelector(0, 1, 0, true)},
		} {
			_, err := NewExclusion([]Selector{tt.sel}, []string{tt.query})
			require.Error(t, err, tt.query)
			assert.Contains(t, err.Error(), tt.query)
		}
	})

	t.Run("実行時のエラーにもクエリが載る", func(t *testing.T) {
		e, err := NewExclusion([]Selector{NewIndexSelector(1), NewRangeSelector(5, 1, 1, false)}, []string{"1", "5:1"})
		require.NoError(t, err)

		_, err = e.Apply(&testColumns{a: []string{"a", "b", "c"}})
		require.Error(t, err)

		xerr, ok := errors.AsType[*ExcludeError](err)
		require.True(t, ok)
		assert.Equal(t, "5:1", xerr.Query)
	})
}
