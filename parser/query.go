package parser

import (
	"regexp"
)

// Query はクエリ文字列を表すやつ
type Query string
type QuerySlice []Query

// start:stop:step
// start:stop
// start
var indexQueryValidator = regexp.MustCompile(`^(-?\d*)(:(-?\d*))?(:(-?\d*))?$`)

// startIndex:/end regexp/
// /start regexp/:endIndex
var switchQueryValidator = regexp.MustCompile(`^(\d+|/.+/):(\+?\d+|/.+/)$`)

func (q Query) isIndexQuery() bool {
	return indexQueryValidator.MatchString(string(q))
}

func (q Query) isSwitchQuery() bool {
	return switchQueryValidator.MatchString(string(q))
}
