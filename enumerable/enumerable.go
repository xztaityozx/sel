package enumerable

import (
	"fmt"
	"strconv"
	"strings"
)

type Range struct {
	start int
	stop  index
	step  int
}

type index struct {
	num int
	inf bool
}

func NewPositionIndex(idx int) index {
	return index{num: idx, inf: false}
}

func NewInfIndex() index {
	return index{inf: true}
}

func NewRange(query string) (Range, error) {
	var err error
	split := strings.Split(query, ":")
	if len(split) == 1 {
		// \d+
		idx, err := strconv.Atoi(split[0])
		return Range{
			start: idx,
			stop:  NewPositionIndex(idx),
			step:  1,
		}, err
	} else if len(split) == 2 || len(split) == 3 {
		// \d*:\d*:\d*
		start := 1
		if len(split[0]) != 0 {
			idx, err := strconv.Atoi(split[0])
			if err != nil {
				return Range{}, err
			}
			start = idx
		}

		stop := NewInfIndex()
		if len(split[1]) != 0 {
			idx, err := strconv.Atoi(split[1])
			if err != nil {
				return Range{}, err
			}
			stop = NewPositionIndex(idx)
		}

		step := 1
		if len(split) == 3 && len(split[2]) != 0 {
			step, err = strconv.Atoi(split[2])
			if err != nil {
				return Range{}, err
			}
		}

		if step == 0 {
			return Range{}, fmt.Errorf("step cannot be zero")
		}

		return Range{start: start, stop: stop, step: step}, nil
	}

	return Range{}, fmt.Errorf("failed to parse query: %s", query)
}

func (r Range) Enumerate(max int) <-chan int {
	rt := make(chan int)

	go func() {
		defer close(rt)

		start := r.start
		if start < 0 {
			start = max + start
		}

		stop := r.stop.num
		if r.stop.inf || stop >= max {
			stop = max
		}
		if stop < 0 {
			stop = max + stop + 1
		}

		step := r.step

		if start == stop {
			rt <- start
		} else if start < stop {
			if step < 0 {
				return
			}
			for idx := start; idx <= stop; idx += step {
				rt <- idx
			}
		} else {
			if step > 0 {
				return
			}
			for idx := start; idx >= stop; idx += step {
				rt <- idx
			}
		}

	}()

	return rt
}
