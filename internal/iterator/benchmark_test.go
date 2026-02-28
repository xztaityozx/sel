package iterator

import (
	"regexp"
	"strings"
	"testing"
)

var testLine = strings.Repeat("column ", 100) // 100 columns

func BenchmarkIterator_ElementAt_First(b *testing.B) {
	iter := NewIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(1)
	}
}

func BenchmarkIterator_ElementAt_Middle(b *testing.B) {
	iter := NewIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(50)
	}
}

func BenchmarkIterator_ElementAt_Last(b *testing.B) {
	iter := NewIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(100)
	}
}

func BenchmarkIterator_ElementAt_Negative(b *testing.B) {
	iter := NewIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(-1)
	}
}

func BenchmarkIterator_ToArray(b *testing.B) {
	iter := NewIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = iter.ToArray()
	}
}

func BenchmarkPreSplitIterator_ElementAt_First(b *testing.B) {
	iter := NewPreSplitIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(1)
	}
}

func BenchmarkPreSplitIterator_ElementAt_Middle(b *testing.B) {
	iter := NewPreSplitIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(50)
	}
}

func BenchmarkPreSplitIterator_ElementAt_Last(b *testing.B) {
	iter := NewPreSplitIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(100)
	}
}

func BenchmarkPreSplitIterator_ElementAt_Negative(b *testing.B) {
	iter := NewPreSplitIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(-1)
	}
}

func BenchmarkPreSplitIterator_ToArray(b *testing.B) {
	iter := NewPreSplitIterator(testLine, " ", false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = iter.ToArray()
	}
}

var regexpSep = regexp.MustCompile(`\s+`)

func BenchmarkRegexpIterator_ElementAt_First(b *testing.B) {
	iter := NewRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(1)
	}
}

func BenchmarkRegexpIterator_ElementAt_Middle(b *testing.B) {
	iter := NewRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(50)
	}
}

func BenchmarkRegexpIterator_ElementAt_Last(b *testing.B) {
	iter := NewRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(100)
	}
}

func BenchmarkRegexpIterator_ElementAt_Negative(b *testing.B) {
	iter := NewRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(-1)
	}
}

func BenchmarkRegexpIterator_ToArray(b *testing.B) {
	iter := NewRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = iter.ToArray()
	}
}

func BenchmarkPreSplitByRegexpIterator_ElementAt_First(b *testing.B) {
	iter := NewPreSplitByRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(1)
	}
}

func BenchmarkPreSplitByRegexpIterator_ElementAt_Middle(b *testing.B) {
	iter := NewPreSplitByRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_, _ = iter.ElementAt(50)
	}
}

func BenchmarkPreSplitByRegexpIterator_ToArray(b *testing.B) {
	iter := NewPreSplitByRegexpIterator(testLine, regexpSep, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter.Reset(testLine)
		_ = iter.ToArray()
	}
}

// Compare strings.Split vs regexp.Split
func BenchmarkStringsSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strings.Split(testLine, " ")
	}
}

func BenchmarkRegexpSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = regexpSep.Split(testLine, -1)
	}
}

// Benchmark map allocation patterns
func BenchmarkMapAllocation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make(map[int]string, 20)
	}
}

func BenchmarkSliceAllocation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]string, 0, 100)
	}
}
