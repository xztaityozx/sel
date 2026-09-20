package column

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/option"
	"github.com/xztaityozx/sel/internal/output"
)

func TestNewRangeSelector(t *testing.T) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	start := rand.Int()
	step := rand.Int()
	stop := rand.Int()

	actual := NewRangeSelector(start, step, stop, true)

	assert.Equal(t, start, actual.start)
	assert.Equal(t, stop, actual.stop)
	assert.Equal(t, step, actual.step)
	assert.True(t, actual.isInfStop)
}

func TestRangeSelector_Select(t *testing.T) {
	var cols []string
	for i := range 20 {
		cols = append(cols, strconv.Itoa(i))
	}

	expectFactory := func(list []int) []string {
		var rt []string
		for _, v := range list {
			rt = append(rt, cols[v])
		}
		return rt
	}

	var buf []byte
	w := bytes.NewBuffer(buf)

	t.Run("OK", func(t *testing.T) {
		dataset := []struct {
			start   int
			step    int
			stop    int
			expects []int
		}{
			{start: 1, step: 1, stop: 5, expects: []int{0, 1, 2, 3, 4}},
			{start: 5, step: -1, stop: 1, expects: []int{4, 3, 2, 1, 0}},
			{start: 1, step: 1, stop: 1, expects: []int{0}},
			{start: -1, step: -1, stop: -5, expects: []int{19, 18, 17, 16, 15}},
			{start: 1, step: 2, stop: 10, expects: []int{0, 2, 4, 6, 8}},
			{start: -1, step: -2, stop: -10, expects: []int{19, 17, 15, 13, 11}},
		}

		for _, v := range dataset {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			expect := expectFactory(v.expects)
			writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, w, true)
			err := rs.Select(writer, &testColumns{a: cols})
			assert.NoError(t, writer.Flush())
			assert.NoError(t, err)
			assert.Equal(t, strings.Join(expect, " "), w.String(), "start: %d, step: %d, stop: %d", v.start, v.step, v.stop)
			w.Reset()
		}
	})

	t.Run("NG", func(t *testing.T) {
		dataset := []struct {
			start int
			step  int
			stop  int
		}{
			{start: 0, step: -1, stop: 5},
			{start: 5, step: 1, stop: 0},
			{start: 1000, step: 1, stop: 1000},
		}

		for _, v := range dataset {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, w, true)
			err := rs.Select(writer, &testColumns{a: cols})
			require.NoError(t, writer.Flush())
			require.Error(t, err)
			assert.Empty(t, w.String())
			w.Reset()
		}
	})

	t.Run("空行はエラーにせず何も書かない", func(t *testing.T) {
		for _, rs := range []RangeSelector{
			NewRangeSelector(1, 1, 1, true),
			NewRangeSelector(1, 1, 3, false),
			NewRangeSelector(-1, -1, -3, false),
		} {
			writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, w, true)
			err := rs.Select(writer, &testColumns{})
			assert.NoError(t, writer.Flush())
			assert.NoError(t, err)
			assert.Empty(t, w.String())
			w.Reset()
		}
	})

	// 桁溢れした i が負に回り込んで行内に戻ってくると、選ぶはずのないカラムを選んでしまう
	t.Run("行幅より大きいstepは1回しか進まない", func(t *testing.T) {
		for _, v := range []struct {
			start   int
			step    int
			stop    int
			expects []int
		}{
			{start: 1, step: math.MaxInt64, stop: 20, expects: []int{0}},
			{start: math.MinInt64, step: math.MaxInt64, stop: 20, expects: []int{19}},
			{start: 20, step: math.MinInt64, stop: 1, expects: []int{19}},
		} {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, w, true)
			err := rs.Select(writer, &testColumns{a: cols})
			require.NoError(t, writer.Flush())
			require.NoError(t, err)
			assert.Equal(t, strings.Join(expectFactory(v.expects), " "), w.String(), "start: %d, step: %d, stop: %d", v.start, v.step, v.stop)
			w.Reset()
		}
	})

	t.Run("Inf", func(t *testing.T) {
		rs := NewRangeSelector(1, 1, 1, true)
		writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, w, true)
		err := rs.Select(writer, &testColumns{a: cols})
		assert.NoError(t, writer.Flush())
		assert.NoError(t, err)
		assert.Equal(t, strings.Join(cols, " "), w.String())
	})
}

func BenchmarkRangeSelector_Select_Forward(b *testing.B) {
	var cols []string
	for i := range 100 {
		cols = append(cols, strconv.Itoa(i))
	}
	rs := NewRangeSelector(1, 1, 100, false)
	opt := option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}
	writer := output.NewWriter(opt, io.Discard, false)
	b.ResetTimer()
	for range b.N {
		_ = rs.Select(writer, &testColumns{a: cols})
		_ = writer.WriteNewLine()
	}
}

func BenchmarkRangeSelector_Select_Backward(b *testing.B) {
	var cols []string
	for i := range 100 {
		cols = append(cols, strconv.Itoa(i))
	}
	rs := NewRangeSelector(100, -1, 1, false)
	opt := option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}
	writer := output.NewWriter(opt, io.Discard, false)
	b.ResetTimer()
	for range b.N {
		_ = rs.Select(writer, &testColumns{a: cols})
		_ = writer.WriteNewLine()
	}
}

func BenchmarkRangeSelector_Select_Step(b *testing.B) {
	var cols []string
	for i := range 100 {
		cols = append(cols, strconv.Itoa(i))
	}
	rs := NewRangeSelector(1, 3, 100, false)
	opt := option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}
	writer := output.NewWriter(opt, io.Discard, false)
	b.ResetTimer()
	for range b.N {
		_ = rs.Select(writer, &testColumns{a: cols})
		_ = writer.WriteNewLine()
	}
}

func TestRangeSelector_Select_OutOfRange(t *testing.T) {
	// 行の外を指す添字は、step の刻みを保ったまま行内に詰める。
	// 詰めないと columns[i-1] が範囲外アクセスになって panic したり、頼んでいない index 0 (行全体) が出たりする
	dataset := []struct {
		name    string
		rs      RangeSelector
		columns []string
		want    string
		wantErr bool
	}{
		{name: "-3:", rs: NewRangeSelector(-3, 1, -3, true), columns: []string{"a"}, want: "a"},
		{name: "1:-5:-1", rs: NewRangeSelector(1, -1, -5, false), columns: []string{"a"}, want: "a"},
		{name: "-4:", rs: NewRangeSelector(-4, 1, -4, true), columns: []string{"a", "b"}, want: "a b"},
		{name: "-8:-1:2", rs: NewRangeSelector(-8, 2, -1, false), columns: []string{"a", "b"}, want: "a"},
		{name: "10:1:-1", rs: NewRangeSelector(10, -1, 1, false), columns: []string{"a", "b", "c"}, want: "c b a"},
		{name: "5::-1", rs: NewRangeSelector(5, -1, 5, true), columns: []string{"a", "b", "c"}, want: "c b a"},
		{name: "-4:-4", rs: NewRangeSelector(-4, 1, -4, false), columns: []string{"a", "b", "c"}, wantErr: true},
		{name: "2:-8:-1", rs: NewRangeSelector(2, -1, -8, false), columns: []string{"a", "b", "c"}, want: "b a"},
		{name: "10:12", rs: NewRangeSelector(10, 1, 12, false), columns: []string{"a", "b", "c"}, want: ""},
		{name: "2:", rs: NewRangeSelector(2, 1, 2, true), columns: []string{"a"}, want: ""},
		{name: "0:0", rs: NewRangeSelector(0, 1, 0, false), columns: []string{"a", "b", "c"}, want: "a b c"},
		{name: "1:3 (空行)", rs: NewRangeSelector(1, 1, 3, false), columns: nil, want: ""},
	}

	for _, v := range dataset {
		t.Run(v.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := output.NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, &buf, true)
			err := v.rs.Select(writer, &testColumns{a: v.columns})
			if v.wantErr {
				assert.ErrorIs(t, err, iterator.ErrIndexOutOfRange)
				return
			}
			require.NoError(t, err)
			require.NoError(t, writer.Flush())
			assert.Equal(t, v.want, buf.String())
		})
	}
}
