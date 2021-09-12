package parser

import (
	"fmt"
	"github.com/xztaityozx/sel/column"
	"strconv"
)

func Parse(queries QuerySlice) ([]column.Selector, error) {
	var rt []column.Selector
	var err error
	for _, query := range queries {
		if query.isIndexQuery() {
			querySection := indexQueryValidator.Split(string(query), -1)
			if len(querySection) == 1 {
				idx, err := strconv.Atoi(querySection[0])
				if err != nil {
					return nil, err
				}
				rt = append(rt, column.NewIndexSelector(idx))
			} else if len(querySection) == 2 || len(querySection) == 3 {
				start := 1
				if len(querySection[0]) != 0 {
					idx, err := strconv.Atoi(querySection[0])
					if err != nil {
						return nil, err
					}
					start = idx
				}

				isInfStop := true
				stop := start
				if len(querySection[1]) != 0 {
					idx, err := strconv.Atoi(querySection[1])
					if err != nil {
						return nil, err
					}
					stop = idx
					isInfStop = false
				}

				step := 1
				if len(querySection) == 3 && len(querySection[2]) != 0 {
					step, err = strconv.Atoi(querySection[2])
					if err != nil {
						return nil, err
					}
				}

				if step == 0 {
					return nil, fmt.Errorf("step cannot be zero")
				}

				rt = append(rt, column.NewRangeSelector(start, step, stop, isInfStop))
			} else {
				return nil, fmt.Errorf("%s is invalid index query", query)
			}
		} else if query.isSwitchQuery() {
			querySection := switchQueryValidator.Split(string(query), -1)
			ss, err := column.NewSwitchSelector(querySection[0], querySection[1])
			if err != nil {
				return nil, err
			}
			rt = append(rt, ss)
		} else {
			return nil, fmt.Errorf("%s is invalid query", query)
		}
	}

	return rt, nil
}
