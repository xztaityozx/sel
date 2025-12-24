package column

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/option"
	"github.com/xztaityozx/sel/internal/output"
)

var testLine = strings.Repeat("column ", 100) // 100 columns

func newTestWriter() *output.Writer {
	opt := option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: " ",
		},
	}
	return output.NewWriter(opt, &bytes.Buffer{}, false)
}

func BenchmarkIndexSelector_First(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewIndexSelector(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkIndexSelector_Middle(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewIndexSelector(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkIndexSelector_Last(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewIndexSelector(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkIndexSelector_Negative(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewIndexSelector(-1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkIndexSelector_Zero(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewIndexSelector(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkRangeSelector_Small(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewRangeSelector(1, 1, 10, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkRangeSelector_Large(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewRangeSelector(1, 1, 100, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkRangeSelector_WithStep(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewRangeSelector(1, 2, 100, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkRangeSelector_Infinite(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel := NewRangeSelector(1, 1, 1, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkSwitchSelector_Regexp(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel, _ := NewSwitchSelector("/col/", "/col/")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkSwitchSelector_Index(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel, _ := NewSwitchSelector("10", "20")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}

func BenchmarkSwitchSelector_AroundContext(b *testing.B) {
	iter := iterator.NewPreSplitIterator(testLine, " ", false)
	w := newTestWriter()
	sel, _ := NewSwitchSelector("/col/", "+5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = sel.Select(w, iter)
	}
}
