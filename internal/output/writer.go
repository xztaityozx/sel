package output

import (
	"bufio"
	"fmt"
	"io"

	"github.com/xztaityozx/sel/internal/option"
)

type Writer struct {
	delimiter      []byte
	buf            *bufio.Writer
	autoFlush      bool
	writtenColumns int
	// template は --template が指定されているときだけ非 nil。
	// このときカラムはデリミタではなくテンプレートのリテラルで繋がれる
	template *option.Template
	// pending は組み立て中の1行分のバイト列。行が完成する(WriteNewLine が呼ばれる)まで
	// buf には書き込まない。こうしておくと、行の途中でエラーが起きて WriteNewLine まで
	// 辿り着けなかったときに、その未完成の断片が buf に混ざらず、buf は常に完成した行だけを
	// 保持する状態になる。呼び出し側はどのタイミングで Flush しても不完全な行が漏れない
	pending []byte
}

var newLine = []byte("\n")

func NewWriter(option option.Option, w io.Writer, autoFlush bool) *Writer {
	return &Writer{
		delimiter: []byte(option.OutPutDelimiter),
		buf:       bufio.NewWriter(w),
		autoFlush: autoFlush,
		template:  option.Template,
	}
}

func (w *Writer) Write(columns ...[]byte) error {
	if len(columns) == 0 {
		return nil
	}

	if w.template != nil {
		// テンプレートのリテラルとカラムを交互に pending へ積んでいく。
		// プレースホルダより多いカラムは書き出さずに捨てるが、数だけは数えておく
		for _, v := range columns {
			if i := w.writtenColumns; i < w.template.Placeholders() {
				w.pending = w.template.AppendColumn(w.pending, i, v)
			}
			w.writtenColumns++
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

// WriteMissing は範囲外だったカラムのぶんを書く。
// fill が空のとき、テンプレートなしなら何も書かない(空カラムを足すと区切り文字だけが増えてしまう)が、
// テンプレートありならプレースホルダを空文字で埋めて数を合わせる
func (w *Writer) WriteMissing(fill []byte) error {
	if len(fill) == 0 && w.template == nil {
		return nil
	}
	return w.Write(fill)
}

// WriteNewLine は改行を書き込んで1行を完成させる。
// テンプレートを利用している場合は、最後のプレースホルダより後ろのリテラルもここで書き込む
func (w *Writer) WriteNewLine() error {
	// ref: Write(columns ...[]byte) error
	if w.template != nil {
		if n := w.template.Placeholders(); w.writtenColumns < n {
			// 埋まらなかったプレースホルダが残っている。組み立て中の行は捨てる
			written := w.writtenColumns
			w.writtenColumns = 0
			w.pending = w.pending[:0]
			return fmt.Errorf("template expects %d columns but query produced %d", n, written)
		}
		w.pending = w.template.AppendTail(w.pending)
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
