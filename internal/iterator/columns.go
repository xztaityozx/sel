package iterator

import (
	"regexp"

	"github.com/xztaityozx/sel/internal/option"
)

// Columns は1行ぶんのカラム列を表す。column.Selector が触るのはこのインターフェースだけ
type Columns interface {
	// ElementAt は idx 番目のカラムを返す。1-indexed で、負の値は末尾からの位置を表す
	ElementAt(idx int) (string, error)
	// ToArray はすべてのカラムを配列にして返す
	ToArray() []string
}

// splitColumns は「1行の文字列」から作り直せる Columns。行を供給する lineSource が使う
type splitColumns interface {
	Columns
	// Reset は s を新しい1行として受け取り、分割状態を作り直す
	Reset(s string)
}

// newSplitColumns は option.Option から適切な splitColumns を生成して返す
func newSplitColumns(option option.Option) (splitColumns, error) {
	if option.UseRegexp {
		r, err := regexp.Compile(option.InputDelimiter)
		if err != nil {
			return nil, err
		}

		if option.SplitBefore {
			// 事前に分割する。選択しないカラムも分割するが、後半のカラムを選択するときにはこちらが有利
			return NewPreSplitByRegexpIterator("", r, option.RemoveEmpty), nil
		}
		// 欲しいところまでの分割を都度行う。前の方にあるindexを選ぶほど有利
		// 負のindexを指定する場合は、末尾まで分割してから返すような実装なので、実行速度が低下してしまうことに注意
		// もしかしたら肯定先読みとか使えば後ろからsplitできたりする？
		return NewRegexpIterator("", r, option.RemoveEmpty), nil
	}

	if option.SplitBefore {
		return NewPreSplitIterator("", option.InputDelimiter, option.RemoveEmpty), nil
	}
	return NewIterator("", option.InputDelimiter, option.RemoveEmpty), nil
}
