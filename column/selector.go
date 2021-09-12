package column

import (
	"errors"
	"fmt"
	"strconv"
)

type Selector interface {
	Select([]string) ([]string, error)
}

// IndexSelector は単一のindexを指定してカラムを選択するやつ
type IndexSelector struct {
	index int
}

func NewIndexSelector(i int) IndexSelector {
	return IndexSelector{index: i}
}

func NewIndexSelectorFromString(str string, def int) (IndexSelector, error) {
	if len(str) == 0 {
		return IndexSelector{index: def}, nil
	}
	num, err := strconv.Atoi(str)
	return NewIndexSelector(num), err
}

func (i IndexSelector) Select(strings []string) ([]string, error) {
	if len(strings) < i.index {
		return nil, errors.New("index out of range")
	}
	if i.index == 0 {
		return strings, nil
	}

	return []string{strings[i.index]}, nil
}

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
			return nil, errors.New("index out of range")
		}
		return []string{strings[start]}, nil
	} else if start < stop {
		if step < 0 {
			return nil, errors.New(fmt.Sprintf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, step, stop))
		}

		var rt []string
		for i := start; i <= stop; i += step {
			rt = append(rt, strings[i])
		}
		return rt, nil
	} else {
		if step > 0 {
			return nil, errors.New(fmt.Sprintf("step must be less than 0(start:step:stop=%d:%d:%d)", start, step, stop))
		}
		var rt []string
		for i := start; i >= stop; i += step {
			rt = append(rt, strings[i])
		}

		return rt, nil
	}
}
