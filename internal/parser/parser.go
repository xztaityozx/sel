package parser

import (
	"fmt"
	"github.com/xztaityozx/sel/internal/column"
	"strconv"
)

func Parse(args []string) ([]column.Selector, error) {
	rt := make([]column.Selector, 0, len(args))
	for _, v := range args {
		query := Query(v)

		if m := query.matchIndexQuery(); m != nil {
			var sel column.Selector
			var err error
			if m[2] == "" {
				sel, err = parseIndexQuery(query, m)
			} else {
				sel, err = parseRangeQuery(query, m)
			}
			if err != nil {
				return nil, err
			}
			rt = append(rt, sel)
			continue
		}

		if m := query.matchSwitchQuery(); m != nil {
			sel, err := parseSwitchQuery(query, m)
			if err != nil {
				return nil, err
			}
			rt = append(rt, sel)
			continue
		}

		return nil, fmt.Errorf("%s is invalid query", query)
	}

	return rt, nil
}

// parseIndexQuery は "1" や "-3" のような単一 index クエリをパースする
func parseIndexQuery(query Query, m []string) (column.Selector, error) {
	sel, err := column.NewIndexSelectorFromString(m[1], 0)
	if err != nil {
		return nil, fmt.Errorf("invalid index %q in query %q: %w", m[1], query, err)
	}
	return sel, nil
}

// parseRangeQuery は "1:10", "1:10:2", "-4:" のような range クエリをパースする
func parseRangeQuery(query Query, m []string) (column.Selector, error) {
	start := 1
	if m[1] != "" {
		v, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid start %q in query %q: %w", m[1], query, err)
		}
		start = v
	}

	isInfStop := true
	stop := start
	if m[3] != "" {
		v, err := strconv.Atoi(m[3])
		if err != nil {
			return nil, fmt.Errorf("invalid stop %q in query %q: %w", m[3], query, err)
		}
		stop = v
		isInfStop = false
	}

	step := 1
	if m[4] != "" && m[5] != "" {
		v, err := strconv.Atoi(m[5])
		if err != nil {
			return nil, fmt.Errorf("invalid step %q in query %q: %w", m[5], query, err)
		}
		step = v
	}

	if step == 0 {
		return nil, fmt.Errorf("step cannot be zero in query %q", query)
	}

	return column.NewRangeSelector(start, step, stop, isInfStop), nil
}

// parseSwitchQuery は "/start/:/end/" のような switch (2addr) クエリをパースする
func parseSwitchQuery(query Query, m []string) (column.Selector, error) {
	sel, err := column.NewSwitchSelector(m[1], m[2])
	if err != nil {
		return nil, fmt.Errorf("invalid switch query %q: %w", query, err)
	}
	return sel, nil
}
