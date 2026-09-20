package parser

import (
	"fmt"
	"strconv"

	"github.com/xztaityozx/sel/internal/column"
)

func Parse(args []string) ([]column.Selector, error) {
	rt := make([]column.Selector, 0, len(args))
	for _, v := range args {
		sel, err := parseQuery(Query(v))
		if err != nil {
			return nil, err
		}
		rt = append(rt, sel)
	}

	return rt, nil
}

// ParseExclude は -x に渡されたクエリ列をパースして column.Exclusion を作る。
// 除外に使えるのは index/range クエリだけで、switch クエリと index 0 は column.NewExclusion が弾く
func ParseExclude(args []string) (*column.Exclusion, error) {
	selectors, err := Parse(args)
	if err != nil {
		return nil, err
	}

	return column.NewExclusion(selectors, args)
}

// parseQuery はクエリ1つを Selector にする
func parseQuery(query Query) (column.Selector, error) {
	if m := query.matchIndexQuery(); m != nil {
		if m[2] == "" {
			return parseIndexQuery(query, m)
		}
		return parseRangeQuery(query, m)
	}

	if m := query.matchSwitchQuery(); m != nil {
		return parseSwitchQuery(query, m)
	}

	return nil, fmt.Errorf("%s is invalid query", query)
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

	// start と stop が両方とも負でなければ、範囲の向きは行のカラム数に関係なく決まる。
	// step の向きと食い違っているクエリはどの行でも成立しないので、毎行エラーにせずここで断る
	// (負の指定は行のカラム数で解決するまで向きが決まらないので、Select / markExcluded に任せる)
	if !isInfStop && start >= 0 && stop >= 0 {
		if start < stop && step < 0 {
			return nil, fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d) in query %q", start, step, stop, query)
		}
		if start > stop && step > 0 {
			return nil, fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d) in query %q", start, step, stop, query)
		}
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
