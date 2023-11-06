package column

import (
	"fmt"
	"github.com/xztaityozx/sel/iterator"
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

func (r RangeSelector) Select(w *Writer, iter iterator.IEnumerable) error {
	strings := iter.ToArray()
	max := len(strings)

	start := r.start
	if start < 0 {
		start = max + start + 1
	}

	stop := r.stop
	if r.isInfStop || stop >= max {
		stop = max
	}
	if stop < 0 {
		stop = max + stop + 1
	}

	step := r.step

	if start == stop {
		if start > max {
			return fmt.Errorf("index out of range")
		}

		return w.Write(strings[start-1])
	} else if start < stop {
		if step < 0 {
			return fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}

		l := 0
		for i := start; i <= stop; i += step {
			if i == 0 {
				l += len(strings)
			} else {
				l++
			}
		}

		rt := make([]string, l)
		idx := 0
		for i := start; i <= stop; i += step {
			if i == 0 {
				for _, v := range strings {
					rt[idx] = v
					idx++
				}
			} else {
				rt[idx] = strings[i-1]
				idx++
			}
		}

		return w.Write(rt...)
	} else {
		if step > 0 {
			return fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}

		l := 0
		for i := start; i >= stop; i += step {
			if i == 0 {
				l += len(strings)
			} else {
				l++
			}
		}

		rt := make([]string, l)
		idx := 0
		for i := start; i >= stop; i += step {
			if i == 0 {
				for _, v := range strings {
					rt[idx] = v
					idx++
				}
			} else {
				rt[idx] = strings[i-1]
				idx++
			}
		}

		return w.Write(rt...)
	}
}
