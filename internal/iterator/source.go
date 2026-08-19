package iterator

import (
	"bufio"
	"encoding/csv"
	"io"
	"strings"

	"github.com/xztaityozx/sel/internal/option"
)

// Source は入力から1行(1レコード)ずつ Columns を供給する。
// 終端に達したら io.EOF を返す。
// 返される Columns は次に Next を呼ぶまでのあいだだけ有効で、実体は使い回される
type Source interface {
	Next() (Columns, error)
}

// NewSource は option.Option に従って input から Columns を供給する Source を作る
func NewSource(option option.Option, input io.Reader) (Source, error) {
	if ok, comma := option.IsXsv(); ok {
		// CSV/TSVのときはencoding/csvが分割をしてくれるので、分割済みの配列をそのまま Columns にすればよい
		r := csv.NewReader(input)
		r.Comma = comma

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
	// 読み取り中に発生したエラー。行を返しきってから次の Next で返す
	err error
}

func (l *lineSource) Next() (Columns, error) {
	if l.err != nil {
		return nil, l.err
	}

	line, err := l.reader.ReadString('\n')
	l.err = err

	if len(line) > 0 {
		// 最終行には改行がないこともあるので、あるときだけ落とす
		l.columns.Reset(strings.TrimRight(line, "\n"))
		return l.columns, nil
	}

	if err == nil {
		// ReadString は1文字以上を返すかerrorを返すので、ここには来ないはず。念のため終端にしておく
		err = io.EOF
		l.err = err
	}

	return nil, err
}

// csvSource は encoding/csv で1レコードずつ読む Source
type csvSource struct {
	reader  *csv.Reader
	columns *PreSplitIterator
}

func (c *csvSource) Next() (Columns, error) {
	record, err := c.reader.Read()
	if err != nil {
		return nil, err
	}

	c.columns.resetFromArray(record)
	return c.columns, nil
}
