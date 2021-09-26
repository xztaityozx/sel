package column

import (
	"fmt"
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

func (r RangeSelector) Select(strings []string) ([]string, error) {
	max := len(strings) - 1

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
			return nil, fmt.Errorf("index out of range")
		}
		return []string{strings[start-1]}, nil
	} else if start < stop {
		if step < 0 {
			return nil, fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}

		var rt []string
		for i := start; i <= stop; i += step {
			if i == 0 {
				rt = append(rt, strings...)
			} else {
				rt = append(rt, strings[i-1])
			}
		}
		return rt, nil
	} else {
		if step > 0 {
			return nil, fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}
		var rt []string
		for i := start; i >= stop; i += step {
			if i == 0 {
				rt = append(rt, strings...)
			} else {
				rt = append(rt, strings[i-1])
			}
		}

		return rt, nil
	}
}
