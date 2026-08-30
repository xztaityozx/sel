package output

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
)

func TestNewWriter(t *testing.T) {
	w := &bytes.Buffer{}
	delim := "d"

	actual := NewWriter(option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: delim,
		},
	}, w, true)

	assert.NotNil(t, actual)
	assert.Equal(t, []byte(delim), actual.delimiter)
	assert.NotNil(t, actual.buf)
	assert.True(t, actual.autoFlush)
}

func TestWriter_Write(t *testing.T) {
	type fields struct {
		delimiter []byte
		buf       *bufio.Writer
	}
	type args struct {
		columns [][]byte
	}
	cols := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	buf := &bytes.Buffer{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "errors",
			fields: fields{
				delimiter: []byte("d"),
				buf:       bufio.NewWriter(buf),
			},
			args:    args{columns: cols},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Writer{
				delimiter: tt.fields.delimiter,
				buf:       tt.fields.buf,
			}
			if err := w.Write(tt.args.columns...); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Write だけでは行が未完成なので、まだ buf には何も渡っていないはず
			_ = w.buf.Flush()
			assert.Equal(t, "", buf.String())

			// WriteNewLine で行が完成して、はじめて buf に渡る
			_ = w.WriteNewLine()
			_ = w.buf.Flush()
			assert.Equal(t, "adbdc\n", buf.String())
		})
	}
}

// TestWriter_Write_PartialLineNotFlushed は、行の途中で終わった(WriteNewLine を呼ばなかった)場合に
// その断片が Flush で漏れないことを確認する。cmd.run() がエラー時に Flush する経路があるための保証
func TestWriter_Write_PartialLineNotFlushed(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, buf, false)

	assert.NoError(t, w.Write([]byte("complete")))
	assert.NoError(t, w.WriteNewLine())

	// 2行目は WriteNewLine まで到達しない(選択が失敗した状況を模す)
	assert.NoError(t, w.Write([]byte("partial")))

	assert.NoError(t, w.Flush())
	assert.Equal(t, "complete\n", buf.String())
}

// newTemplateWriter は --template を指定した Writer を作るテストヘルパ
func newTemplateWriter(w io.Writer, tmpl string) *Writer {
	return NewWriter(option.Option{
		DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "},
		Template:        option.ParseTemplate(tmpl),
	}, w, false)
}

func TestWriter_Template(t *testing.T) {
	tests := []struct {
		name     string
		template string
		columns  []string
		want     string
	}{
		{name: "プレースホルダとカラムが交互に並ぶ", template: "one: {} two: {}", columns: []string{"a", "b"}, want: "one: a two: b\n"},
		{name: "先頭と末尾のリテラルなし", template: "{}{}", columns: []string{"a", "b"}, want: "ab\n"},
		{name: "{{ と }} はリテラルの波括弧になる", template: "x{{.}}y {}", columns: []string{"a"}, want: "x{.}y a\n"},
		{name: "{{}} はリテラルの {} になる", template: "{{}} {}", columns: []string{"a"}, want: "{} a\n"},
		{name: "プレースホルダより多いカラムは捨てられる", template: "{}-{}", columns: []string{"a", "b", "c"}, want: "a-b\n"},
		{name: "プレースホルダなしならリテラルだけが出る", template: "no placeholder", columns: nil, want: "no placeholder\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := newTemplateWriter(buf, tt.template)

			for _, c := range tt.columns {
				assert.NoError(t, w.Write([]byte(c)))
			}
			assert.NoError(t, w.WriteNewLine())
			assert.NoError(t, w.Flush())

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestWriter_Template_NotEnoughColumns は、プレースホルダを埋めきれなかったときに
// 自前のエラーを返し、書きかけの行を漏らさないことを確認する
func TestWriter_Template_NotEnoughColumns(t *testing.T) {
	buf := &bytes.Buffer{}
	w := newTemplateWriter(buf, "{} {} {}")

	assert.NoError(t, w.Write([]byte("a")))
	assert.NoError(t, w.Write([]byte("b")))

	err := w.WriteNewLine()
	assert.EqualError(t, err, "template expects 3 columns but query produced 2")

	assert.NoError(t, w.Flush())
	assert.Equal(t, "", buf.String())
}

// TestWriter_Template_MultipleLines は行をまたいでも状態が持ち越されないことを確認する
func TestWriter_Template_MultipleLines(t *testing.T) {
	buf := &bytes.Buffer{}
	w := newTemplateWriter(buf, "[{}:{}]")

	for _, cols := range [][]string{{"a", "b"}, {"c", "d"}} {
		for _, c := range cols {
			assert.NoError(t, w.Write([]byte(c)))
		}
		assert.NoError(t, w.WriteNewLine())
	}
	assert.NoError(t, w.Flush())

	assert.Equal(t, "[a:b]\n[c:d]\n", buf.String())
}

func BenchmarkWriter_Write(b *testing.B) {
	w := NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, io.Discard, false)
	cols := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Write(cols...)
		_ = w.WriteNewLine()
	}
}

func BenchmarkWriter_WriteNewLine_Template(b *testing.B) {
	w := newTemplateWriter(io.Discard, "{} {} {} {} {}")

	cols := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Write(cols...)
		_ = w.WriteNewLine()
	}
}
