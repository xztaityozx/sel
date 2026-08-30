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
			if err := w.Write(columns...); err != nil {
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
			if err := w.Write(columns...); err != nil {
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
