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
	benchmarkPipeline(b, 10, 5)
}

// BenchmarkPipelineWide はカラム数の多い行で後ろの方のカラムを選ぶ。
// front を最後まで伸ばす経路なので、Reset がバッファを使い回せているかがここに出る
func BenchmarkPipelineWide(b *testing.B) {
	benchmarkPipeline(b, 100, 99)
}

// benchmarkPipeline は columns カラムの行を並べた入力を流して idx 番目のカラムを取り出す
func benchmarkPipeline(b *testing.B, columns, idx int) {
	const lines = 200000

	var data bytes.Buffer
	for i := 0; i < lines; i++ {
		for c := 0; c < columns; c++ {
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

			s, err := columns.ElementAt(idx)
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
