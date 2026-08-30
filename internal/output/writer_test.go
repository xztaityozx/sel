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

// TestWriter_WriteMissing は範囲外カラムの埋め方が、テンプレートの有無で
// 期待通りに変わることを確認する。テンプレートなしでは空の fill で区切り文字を増やさず、
// テンプレートありでは空でもプレースホルダを1つ消費する
func TestWriter_WriteMissing(t *testing.T) {
	tests := []struct {
		name     string
		template string
		fill     string
		want     string
	}{
		{name: "テンプレートなし・空の fill は何も足さない", template: "", fill: "", want: "a\n"},
		{name: "テンプレートなし・fill ありは1カラムとして書く", template: "", fill: "x", want: "a x\n"},
		{name: "テンプレートあり・空の fill でもプレースホルダを埋める", template: "1st={} 5th={}", fill: "", want: "1st=a 5th=\n"},
		{name: "テンプレートあり・fill ありはその値で埋める", template: "1st={} 5th={}", fill: "x", want: "1st=a 5th=x\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var w *Writer
			if tt.template == "" {
				w = NewWriter(option.Option{
					DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "},
				}, buf, false)
			} else {
				w = newTemplateWriter(buf, tt.template)
			}

			assert.NoError(t, w.Write([]byte("a")))
			assert.NoError(t, w.WriteMissing([]byte(tt.fill)))
			assert.NoError(t, w.WriteNewLine())
			assert.NoError(t, w.Flush())

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestWriter_WriteLine は index 0 (行全体) の書き込みを確認する。
// 分割結果が複数カラムでもテンプレートではプレースホルダを1つだけ消費すること、
// 空行(カラム0個)は「カラムが0個」ではなく「空文字列のカラムが1個」として扱うことを見る
func TestWriter_WriteLine(t *testing.T) {
	tests := []struct {
		name     string
		template string
		columns  [][]byte
		want     string
	}{
		{name: "テンプレートなし・複数カラムはdelimiterで繋がる", template: "", columns: [][]byte{[]byte("a"), []byte("b")}, want: "a b\n"},
		{name: "テンプレートなし・空行は空行のまま", template: "", columns: nil, want: "\n"},
		{name: "テンプレートあり・複数カラムでもプレースホルダは1つだけ消費する", template: "<{}>", columns: [][]byte{[]byte("a"), []byte("b")}, want: "<a b>\n"},
		{name: "テンプレートあり・空行はプレースホルダ1つを空文字で埋める", template: "<{}>", columns: nil, want: "<>\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var w *Writer
			if tt.template == "" {
				w = NewWriter(option.Option{
					DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "},
				}, buf, false)
			} else {
				w = newTemplateWriter(buf, tt.template)
			}

			assert.NoError(t, w.WriteLine(tt.columns))
			assert.NoError(t, w.WriteNewLine())
			assert.NoError(t, w.Flush())

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestWriter_FillRemaining は、range クエリのようにエラーを出さずに列数が足りないまま
// 終わったときでも、行末で残りのプレースホルダをまとめて埋められることを確認する
func TestWriter_FillRemaining(t *testing.T) {
	tests := []struct {
		name    string
		written []string
		fill    string
		want    string
	}{
		{name: "残りが1個なら1個だけ埋める", written: []string{"a"}, fill: "", want: "[a|]\n"},
		{name: "fillありならその値で埋める", written: []string{"a"}, fill: "x", want: "[a|x]\n"},
		{name: "1つも書かれていなければ全部埋める", written: nil, fill: "x", want: "[x|x]\n"},
		{name: "すでに全部埋まっていれば何もしない", written: []string{"a", "b"}, fill: "x", want: "[a|b]\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			w := newTemplateWriter(buf, "[{}|{}]")

			for _, c := range tt.written {
				assert.NoError(t, w.Write([]byte(c)))
			}
			assert.NoError(t, w.FillRemaining([]byte(tt.fill)))
			assert.NoError(t, w.WriteNewLine())
			assert.NoError(t, w.Flush())

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestWriter_FillRemaining_NoTemplate はテンプレートなしのときに no-op であることを確認する。
// range クエリの列不足はテンプレートなしの出力ではパディングしない、という既存の挙動を守るため
func TestWriter_FillRemaining_NoTemplate(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, buf, false)

	assert.NoError(t, w.Write([]byte("a")))
	assert.NoError(t, w.FillRemaining([]byte("x")))
	assert.NoError(t, w.WriteNewLine())
	assert.NoError(t, w.Flush())

	assert.Equal(t, "a\n", buf.String())
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
