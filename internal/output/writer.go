package output

import (
	"bufio"
	"io"
	"text/template"

	"github.com/xztaityozx/sel/internal/option"
	"github.com/xztaityozx/sel/internal/sliceutil"
)

type Writer struct {
	delimiter      []byte
	buf            *bufio.Writer
	autoFlush      bool
	writtenColumns int
	outputTemplate *template.Template
	column         []string
	// pending は組み立て中の1行分のバイト列。行が完成する(WriteNewLine が呼ばれる)まで
	// buf には書き込まない。こうしておくと、行の途中でエラーが起きて WriteNewLine まで
	// 辿り着けなかったときに、その未完成の断片が buf に混ざらず、buf は常に完成した行だけを
	// 保持する状態になる。呼び出し側はどのタイミングで Flush しても不完全な行が漏れない
	pending []byte
}

var newLine = []byte("\n")

func NewWriter(option option.Option, w io.Writer, autoFlush bool) *Writer {
	return &Writer{
		delimiter:      []byte(option.OutPutDelimiter),
		buf:            bufio.NewWriter(w),
		autoFlush:      autoFlush,
		outputTemplate: option.Template,
		column:         []string{},
	}
}

func (w *Writer) Write(columns ...[]byte) error {
	if len(columns) == 0 {
		return nil
	}

	if w.outputTemplate != nil {
		// テンプレートを使うときは、出力すべきすべてのカラムが揃ってから書き出すので、ここにはバッファに乗せるのみ
		// 実際の書き込みは WriteNewLine() で行う。
		// text/template に渡すために、ここでだけ文字列へコピーする
		for _, v := range columns {
			w.column = append(w.column, string(v))
		}
		return nil
	}

	if w.writtenColumns != 0 {
		w.pending = append(w.pending, w.delimiter...)
	}
	w.pending = append(w.pending, columns[0]...)
	for _, v := range columns[1:] {
		w.pending = append(w.pending, w.delimiter...)
		w.pending = append(w.pending, v...)
	}

	w.writtenColumns += len(columns)

	if w.autoFlush {
		if _, err := w.buf.Write(w.pending); err != nil {
			return err
		}
		w.pending = w.pending[:0]
		return w.buf.Flush()
	}

	return nil
}

// WriteNewLine は改行を書き込む。テンプレートを利用している場合は、テンプレートを使った書き込みを行う
func (w *Writer) WriteNewLine() error {
	// ref: Write(columns ...string) error
	if w.outputTemplate != nil {
		err := w.outputTemplate.Execute(w.buf, w.column)
		if err != nil {
			return err
		}
		w.column = sliceutil.Reset(w.column)
		w.writtenColumns = 0
		_, err = w.buf.Write(newLine)
		return err
	}

	w.writtenColumns = 0
	w.pending = append(w.pending, newLine...)
	_, err := w.buf.Write(w.pending)
	w.pending = w.pending[:0]
	return err
}

func (w *Writer) Flush() error {
	return w.buf.Flush()
}
