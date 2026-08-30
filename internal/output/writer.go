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

// WriteLine は index 0 (awk の $0 相当、行全体) を書く。
// columns は iter.ToArray() の結果をそのまま渡す想定。
// $0 は分割結果が何個あろうと「行全体」という1つの意味的な単位なので、
// テンプレートを使っているときはプレースホルダを1つだけ消費する
// (Write に columns をそのまま渡すと、分割された数だけプレースホルダを消費してしまう)。
//
// テンプレートなしのときは delimiter で繋いで書くだけなので、書き込まれるバイト列は
// Write(columns...) をそのまま呼んだ場合と変わらない。ただし空行(columns の要素数が0)は
// 「カラムが0個」ではなく「空文字列のカラムが1個」として扱う必要がある。
// ToArray() は空入力に対してカラム0個を返す(分割の規則どおり)が、
// そのまま Write に渡すと len(columns)==0 で何もしなくなり、
// テンプレート使用時にプレースホルダを1つも消費できず「produced 0」エラーになってしまうため
func (w *Writer) WriteLine(columns [][]byte) error {
	if w.template == nil {
		if len(columns) == 0 {
			return w.Write([]byte{})
		}
		return w.Write(columns...)
	}

	if i := w.writtenColumns; i < w.template.Placeholders() {
		w.pending = w.template.AppendLiteral(w.pending, i)
		for j, v := range columns {
			if j > 0 {
				w.pending = append(w.pending, w.delimiter...)
			}
			w.pending = append(w.pending, v...)
		}
	}
	w.writtenColumns++
	return nil
}

// FillRemaining はテンプレート使用時に、埋まっていないプレースホルダをすべて fill で埋める。
// IndexSelector の範囲外アクセスは selector.Select がエラーを返すので WriteMissing で個別に埋められるが、
// RangeSelector は列数が足りなくてもエラーにせず黙って少ない列数で書き出す(クランプする)ため、
// そのエラーを起点にした補完ができない。そこで行の終わりにこれを呼び、
// -M/-E が有効なときは残ったプレースホルダをまとめて埋める。
// テンプレートを使っていないときは何もしない(range クエリの列不足は元々エラーではなく、
// テンプレートなしの出力ではパディングもしない、という既存の挙動を変えないため)
func (w *Writer) FillRemaining(fill []byte) error {
	if w.template == nil {
		return nil
	}
	for w.writtenColumns < w.template.Placeholders() {
		if err := w.Write(fill); err != nil {
			return err
		}
	}
	return nil
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
