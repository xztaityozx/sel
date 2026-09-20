package column

import (
	"fmt"

	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
)

// RangeSelector はカラムの範囲選択するやつ
type RangeSelector struct {
	start     int
	step      int
	stop      int
	isInfStop bool
}

func NewRangeSelector(start, step, stop int, isInfStop bool) RangeSelector {
	return RangeSelector{start: start, step: step, stop: stop, isInfStop: isInfStop}
}

func (r RangeSelector) Select(w *output.Writer, iter iterator.Columns) error {
	columns := iter.ToArray()
	m := len(columns)

	if m == 0 {
		// 空行。選べるカラムが1つもないので何も書かない。
		// range クエリは行が短くても少ない数だけ書いてエラーにしない約束なので、その極端な場合として扱う
		// (normalizeRange は stop を 0 に潰してしまい、start > stop = step の向き違いのエラーに化ける)
		return nil
	}

	start, stop, step := r.normalizeRange(m)

	if start == stop {
		if start > m || start < 1 {
			return fmt.Errorf("index %d: %w", start, iterator.ErrIndexOutOfRange)
		}
		return w.Write(columns[start-1])
	}

	if start < stop {
		if step < 0 {
			return fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}
		return r.selectForward(w, columns, start, stop, step)
	}

	// start > stop
	if step > 0 {
		return fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d)", start, step, stop)
	}
	return r.selectBackward(w, columns, start, stop, step)
}

// normalizeRange は範囲パラメータを正規化する
func (r RangeSelector) normalizeRange(m int) (start, stop, step int) {
	start = r.start
	if start < 0 {
		start = m + start + 1
	}

	stop = r.stop
	if r.isInfStop || stop >= m {
		stop = m
	}
	if stop < 0 {
		stop = m + stop + 1
	}

	return start, stop, r.step
}

// selectForward は start < stop の場合の選択処理
func (r RangeSelector) selectForward(w *output.Writer, columns [][]byte, start, stop, step int) error {
	for i := start; i <= stop; i += step {
		if i == 0 {
			// i == 0 は index 0 (行全体) の指定。空行では columns が0個になりうるが、
			// $0 は「カラムが0個」ではなく「空文字列のカラムが1個」として書く(WriteLine 参照)
			if err := w.WriteLine(columns); err != nil {
				return err
			}
		} else {
			if err := w.Write(columns[i-1]); err != nil {
				return err
			}
		}
	}
	return nil
}

// selectBackward は start > stop の場合の選択処理
func (r RangeSelector) selectBackward(w *output.Writer, columns [][]byte, start, stop, step int) error {
	for i := start; i >= stop; i += step {
		if i == 0 {
			// i == 0 は index 0 (行全体) の指定。空行では columns が0個になりうるが、
			// $0 は「カラムが0個」ではなく「空文字列のカラムが1個」として書く(WriteLine 参照)
			if err := w.WriteLine(columns); err != nil {
				return err
			}
		} else {
			if err := w.Write(columns[i-1]); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolve は行のカラム数 n に対して start/stop を実際のカラム番号に解決する。
// normalizeRange と違って stop を n にクランプしない。クランプすると、行より後ろを指す
// 範囲指定(5列の行に対する 10:12 など)が start > stop になって「step の向きが違う」に化けてしまう
func (r RangeSelector) resolve(n int) (start, stop int) {
	start = r.start
	if start < 0 {
		start = n + start + 1
	}

	if r.isInfStop {
		return start, n
	}

	stop = r.stop
	if stop < 0 {
		stop = n + stop + 1
	}

	return start, stop
}

// markExcluded は -x でこの範囲が指すカラムに印をつける。
// 範囲外の添字は黙って飛ばすが、step の向きが範囲と食い違っているのはクエリ自体の誤りなのでエラーにする
// (Select と同じ扱い)
func (r RangeSelector) markExcluded(mark []bool) error {
	n := len(mark)
	start, stop := r.resolve(n)

	// 開いた範囲(2: など)で start が行末を越えているのは、向きの食い違いではなく除外対象なし
	if r.isInfStop && start > stop {
		return nil
	}

	if start <= stop {
		if r.step < 0 {
			return fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, r.step, stop)
		}
		for i := start; i <= stop; i += r.step {
			markColumn(mark, i)
		}
		return nil
	}

	if r.step > 0 {
		return fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d)", start, r.step, stop)
	}
	for i := start; i >= stop; i += r.step {
		markColumn(mark, i)
	}
	return nil
}

// markColumn は行のカラム数に収まっている i だけに印をつける
func markColumn(mark []bool, i int) {
	if i >= 1 && i <= len(mark) {
		mark[i-1] = true
	}
}
