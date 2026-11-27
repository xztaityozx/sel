package column

import (
	"fmt"

	"github.com/xztaityozx/sel/internal/iterator"
)

// excluder は -x に書けるクエリ(index/range)が実装するやつ。
// mark は行のカラム数ぶんの長さで、除外するカラム(1-indexed の i)について mark[i-1] を立てる。
// 存在しないカラムを指す指定は何もしない(除外すべき列がないだけでエラーではない)
type excluder interface {
	markExcluded(mark []bool)
}

// Exclusion は -x で指定されたカラムを取り除いた Columns を作るやつ。
// 除外は選択より先に行われ、残ったカラムは1から番号付けし直される
type Exclusion struct {
	excluders []excluder
	// 行ごとに使い回すバッファ。backing array を手放さない理由は iterator.Iterator.Reset と同じ
	mark []bool
	kept [][]byte
	view *iterator.PreSplitIterator
}

// NewExclusion は -x のクエリをパースした Selector から Exclusion を作る。
// queries は selectors と同じ順番のクエリ文字列で、エラーメッセージに使う
func NewExclusion(selectors []Selector, queries []string) (*Exclusion, error) {
	e := &Exclusion{
		excluders: make([]excluder, 0, len(selectors)),
		view:      iterator.NewArrayColumns(),
	}

	for i, s := range selectors {
		query := ""
		if i < len(queries) {
			query = queries[i]
		}

		// index 0 (行全体) は除外できない。range に埋まっている 0 も同じ
		switch q := s.(type) {
		case IndexSelector:
			if q.index == 0 {
				return nil, fmt.Errorf("query %q: cannot exclude index 0 (whole line)", query)
			}
		case RangeSelector:
			if q.includesWholeLine() {
				return nil, fmt.Errorf("query %q: cannot exclude index 0 (whole line)", query)
			}
		}

		x, ok := s.(excluder)
		if !ok {
			return nil, fmt.Errorf("query %q: only index and range queries can be excluded", query)
		}
		e.excluders = append(e.excluders, x)
	}

	return e, nil
}

// Apply は除外後のカラムだけを見せる Columns を返す。
// 返される Columns とそこから取り出した []byte は、次に Apply を呼ぶまでのあいだだけ有効
// (iterator.Source が定めている寿命と同じ)
func (e *Exclusion) Apply(columns iterator.Columns) iterator.Columns {
	a := columns.ToArray()

	if cap(e.mark) < len(a) {
		e.mark = make([]bool, len(a))
	} else {
		e.mark = e.mark[:len(a)]
		clear(e.mark)
	}

	for _, x := range e.excluders {
		x.markExcluded(e.mark)
	}

	e.kept = e.kept[:0]
	for i, v := range a {
		if !e.mark[i] {
			e.kept = append(e.kept, v)
		}
	}

	e.view.ResetFromArray(e.kept)
	return e.view
}
