package iterator

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/xztaityozx/sel/internal/option"
	"github.com/xztaityozx/sel/internal/output"
)

// BenchmarkPipeline は Source → Columns → output.Writer という実行時と同じ経路を回す。
// 個々の Iterator ではなくパイプライン全体を見るので、行あたりのアロケーションの回帰を検出できる
func BenchmarkPipeline(b *testing.B) {
	const lines = 200000

	var data bytes.Buffer
	for i := 0; i < lines; i++ {
		for c := 0; c < 10; c++ {
			if c > 0 {
				data.WriteByte(' ')
			}
			fmt.Fprintf(&data, "col%d_%07d", c+1, i)
		}
		data.WriteByte('\n')
	}

	opt := option.Option{
		DelimiterOption: option.DelimiterOption{
			InputDelimiter:  " ",
			OutPutDelimiter: " ",
		},
	}

	b.ReportAllocs()
	b.SetBytes(int64(data.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src, err := NewSource(opt, bytes.NewReader(data.Bytes()))
		if err != nil {
			b.Fatal(err)
		}

		w := output.NewWriter(opt, io.Discard, false)
		for {
			columns, err := src.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				b.Fatal(err)
			}

			s, err := columns.ElementAt(5)
			if err != nil {
				b.Fatal(err)
			}
			if err := w.Write(s); err != nil {
				b.Fatal(err)
			}
			if err := w.WriteNewLine(); err != nil {
				b.Fatal(err)
			}
		}

		if err := w.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}
