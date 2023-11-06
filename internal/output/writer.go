package output

import (
	"bufio"
	"github.com/xztaityozx/sel/internal/option"
	"io"
	"text/template"
)

type Writer struct {
	delimiter      []byte
	buf            *bufio.Writer
	autoFlush      bool
	writtenColumns int
	outputTemplate *template.Template
	column         []string
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

func (w *Writer) Write(columns ...string) error {
	if len(columns) == 0 {
		return nil
	}

	if w.outputTemplate != nil {
		// テンプレートを使うときは、出力すべきすべてのカラムが揃ってから書き出すので、ここにはバッファに乗せるのみ
		// 実際の書き込みは WriteNewLine() で行う
		w.column = append(w.column, columns...)
		return nil
	}

	if w.writtenColumns != 0 {
		if _, err := w.buf.Write(w.delimiter); err != nil {
			return err
		}
	}

	if _, err := w.buf.WriteString(columns[0]); err != nil {
		return err
	}

	for _, v := range columns[1:] {
		if _, err := w.buf.Write(w.delimiter); err != nil {
			return err
		}
		if _, err := w.buf.WriteString(v); err != nil {
			return err
		}
	}

	w.writtenColumns += len(columns)

	if w.autoFlush {
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
		w.column = []string{}

		return nil
	}

	w.writtenColumns = 0
	_, err := w.buf.Write(newLine)
	return err
}

func (w *Writer) Flush() error {
	return w.buf.Flush()
}
