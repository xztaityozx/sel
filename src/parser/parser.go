package parser

import (
	"fmt"
	"github.com/xztaityozx/sel/src/column"
	"strconv"
	"strings"
)

func Parse(args []string) ([]column.Selector, error) {
	queries := make(QuerySlice, 0, len(args))
	for _, v := range args {
		queries = append(queries, Query(v))
	}

	rt := make([]column.Selector, 0, len(args))
	var err error
	for _, query := range queries {
		if query.isIndexQuery() {
			querySection := strings.Split(string(query), ":")
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
			// sedやawkの2addrみたいなやつ
			// /regexp/,/regexp/
			// /regexp/,number
			// number,/regexp/
			s := switchQueryValidator.FindAllStringSubmatch(string(query), -1)[0]
			ss, err := column.NewSwitchSelector(s[1], s[2])
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
