package output

import (
	"bytes"
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
		_ = w.Write([]byte("column"))
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
	columns := bytes.Split(bytes.Repeat([]byte("column "), 100), []byte(" "))
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
		_ = w.Write([]byte("column"))
		_ = w.WriteNewLine()
	}
}
