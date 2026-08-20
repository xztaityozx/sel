package iterator

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"

	"github.com/xztaityozx/sel/internal/option"
)

// Source は入力から1行(1レコード)ずつ Columns を供給する。
// 終端に達したら io.EOF を返す。
// 返される Columns は次に Next を呼ぶまでのあいだだけ有効で、実体は使い回される。
// Columns から取り出した []byte も同じ寿命しかない(読み取りバッファをそのまま指しているため)ので、
// 次の Next をまたいで保持したいときは bytes.Clone などでコピーすること
type Source interface {
	Next() (Columns, error)
}

// NewSource は option.Option に従って input から Columns を供給する Source を作る
func NewSource(option option.Option, input io.Reader) (Source, error) {
	if ok, comma := option.IsXsv(); ok {
		// CSV/TSVのときはencoding/csvが分割をしてくれるので、分割済みの配列をそのまま Columns にすればよい
		r := csv.NewReader(input)
		r.Comma = comma

		// レコードごとに []string を作り直さないようにする。中身は Next で自前のバッファに写す
		r.ReuseRecord = true

		return &csvSource{
			reader:  r,
			columns: NewPreSplitIterator("", string(comma), option.RemoveEmpty),
		}, nil
	}

	columns, err := newSplitColumns(option)
	if err != nil {
		return nil, err
	}

	return &lineSource{
		reader:  bufio.NewReader(input),
		columns: columns,
	}, nil
}

// lineSource は1行ずつ読んで splitColumns に渡す Source
type lineSource struct {
	reader  *bufio.Reader
	columns splitColumns
	// bufio のバッファに収まらない長い行を組み立てるための一時バッファ
	buf []byte
	// 読み取り中に発生したエラー。行を返しきってから次の Next で返す
	err error
}

func (l *lineSource) Next() (Columns, error) {
	if l.err != nil {
		return nil, l.err
	}

	line, err := l.readLine()
	l.err = err

	if len(line) > 0 {
		// bufio のバッファ上のスライスをそのまま渡す。コピーしないので1行あたりのアロケーションが消える。
		// このスライス(とそこから切り出すカラム)は次に readLine を呼ぶまでしか有効でないが、
		// Source が「返した Columns は次の Next までのあいだだけ有効」と定めているのと同じ寿命なので問題ない
		// 最終行には改行がないこともあるので、あるときだけ落とす
		l.columns.Reset(trimNewline(line))
		return l.columns, nil
	}

	if err == nil {
		// readLine は1バイト以上を返すかerrorを返すので、ここには来ないはず。念のため終端にしておく
		err = io.EOF
		l.err = err
	}

	return nil, err
}

// readLine は改行までを bufio のバッファ上のスライスとして返す。
// 返されたスライスは次に readLine を呼ぶまでのあいだだけ有効
func (l *lineSource) readLine() ([]byte, error) {
	line, err := l.reader.ReadSlice('\n')
	if !errors.Is(err, bufio.ErrBufferFull) {
		return line, err
	}

	// バッファに収まらない行のときだけ、l.buf に連結して1行にする
	l.buf = append(l.buf[:0], line...)
	for errors.Is(err, bufio.ErrBufferFull) {
		line, err = l.reader.ReadSlice('\n')
		l.buf = append(l.buf, line...)
	}

	return l.buf, err
}

// trimNewline は末尾の改行を落とす
func trimNewline(b []byte) []byte {
	for len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return b
}

// csvSource は encoding/csv で1レコードずつ読む Source
type csvSource struct {
	reader  *csv.Reader
	columns *PreSplitIterator
	// レコードを写しておくバッファ。レコードごとに使い回す
	buf     []byte
	offsets []int
	fields  [][]byte
}

func (c *csvSource) Next() (Columns, error) {
	record, err := c.reader.Read()
	if err != nil {
		return nil, err
	}

	c.columns.resetFromArray(c.toFields(record))
	return c.columns, nil
}

// toFields は encoding/csv が返す []string を、使い回しのバッファ上の [][]byte に写す。
// 連結してから切り出すのは、append によるバッファの作り直しで先頭のスライスが無効になるのを避けるため
func (c *csvSource) toFields(record []string) [][]byte {
	c.buf = c.buf[:0]
	c.offsets = c.offsets[:0]
	for _, f := range record {
		c.buf = append(c.buf, f...)
		c.offsets = append(c.offsets, len(c.buf))
	}

	c.fields = c.fields[:0]
	beg := 0
	for _, end := range c.offsets {
		c.fields = append(c.fields, c.buf[beg:end])
		beg = end
	}

	return c.fields
}
