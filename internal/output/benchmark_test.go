package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xztaityozx/sel/internal/option"
)

func BenchmarkWriter_Write_Single(b *testing.B) {
	opt := option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: " ",
		},
	}
	buf := &bytes.Buffer{}
	w := NewWriter(opt, buf, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = w.Write("column")
		_ = w.WriteNewLine()
	}
}

func BenchmarkWriter_Write_Multiple(b *testing.B) {
	opt := option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: " ",
		},
	}
	buf := &bytes.Buffer{}
	w := NewWriter(opt, buf, false)
	columns := strings.Split(strings.Repeat("column ", 100), " ")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = w.Write(columns...)
		_ = w.WriteNewLine()
	}
}

func BenchmarkWriter_Write_WithFlush(b *testing.B) {
	opt := option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: " ",
		},
	}
	buf := &bytes.Buffer{}
	w := NewWriter(opt, buf, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = w.Write("column")
		_ = w.WriteNewLine()
	}
}
