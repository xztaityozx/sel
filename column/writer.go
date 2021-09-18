package column

import (
	"bufio"
	"io"
)

type Writer struct {
	delimiter []byte
	buf       *bufio.Writer
	autoFlush bool
}

func NewWriter(delimiter string, w io.Writer) *Writer {
	return &Writer{delimiter: []byte(delimiter), buf: bufio.NewWriter(w), autoFlush: false}
}

func (w *Writer) SetAutoFlush(b bool) {
	w.autoFlush = b
}

func (w *Writer) Write(columns []string) error {
	if len(columns) == 0 {
		return nil
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

	if w.autoFlush {
		return w.buf.Flush()
	}

	return nil
}
