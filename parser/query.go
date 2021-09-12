package parser

import "regexp"

//go:generate genny -in "../gen/collections.go" -out "./gen-query-collections.go" -pkg "parser" gen "TSource=Query TResult=Index"
// Query はクエリ文字列を表すやつ
type Query string

var indexQueryValidator = regexp.MustCompile(`-?\d*(:-?\d*(:-?\d*)?)?`)
var switchQueryValidator = regexp.MustCompile(`(\d+|/.+/):(\+?\d+|/.+/)`)

func (q Query) isIndexQuery() bool {
	return indexQueryValidator.MatchString(string(q))
}

func (q Query) isSwitchQuery() bool {
	return switchQueryValidator.MatchString(string(q))
}
