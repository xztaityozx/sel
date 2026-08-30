package parser

import (
	"regexp"
)

// Query はクエリ文字列を表すやつ
type Query string

// start:stop:step
// start:stop
// start
var indexQueryValidator = regexp.MustCompile(`^(-?\d*)(:(-?\d*))?(:(-?\d*))?$`)

// startIndex:/end regexp/
// /start regexp/:endIndex
var switchQueryValidator = regexp.MustCompile(`^(\d+|/.+/):(\+?\d+|/.+/)$`)

// matchIndexQuery はクエリが index/range クエリの形式にマッチするか判定し、
// マッチした場合はサブマッチ結果を返す(マッチしなければ nil)
func (q Query) matchIndexQuery() []string {
	return indexQueryValidator.FindStringSubmatch(string(q))
}

// matchSwitchQuery はクエリが switch クエリの形式にマッチするか判定し、
// マッチした場合はサブマッチ結果を返す(マッチしなければ nil)
func (q Query) matchSwitchQuery() []string {
	return switchQueryValidator.FindStringSubmatch(string(q))
}
