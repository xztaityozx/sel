package iterator

import (
	"bytes"
	"errors"
	"regexp"
)

// PreSplitIterator は先に全カラムへ分割してしまう Columns
type PreSplitIterator struct {
	a           [][]byte
	sep         []byte
	reg         *regexp.Regexp
	l           int
	removeEmpty bool
}

func (p *PreSplitIterator) ElementAt(idx int) ([]byte, error) {
	if p.l < idx {
		return nil, errors.New(IndexOutOfRange)
	}

	if idx < 0 {
		if -p.l > idx {
			return nil, errors.New(IndexOutOfRange)
		}
		return p.a[p.l+idx], nil
	}

	return p.a[idx-1], nil
}

func (p *PreSplitIterator) ToArray() [][]byte {
	return p.a
}

func (p *PreSplitIterator) Reset(b []byte) {
	if p.reg == nil {
		p.resetFromArray(bytes.Split(b, p.sep))
	} else {
		p.resetFromArray(splitByRegexp(p.reg, b))
	}
}

// resetFromArray は分割済みの配列をそのままカラム列として受け取る。
// encoding/csv のように分割済みのレコードが手に入る入力で使う
func (p *PreSplitIterator) resetFromArray(a [][]byte) {
	if p.removeEmpty {
		p.a = removeEmpty(a)
	} else {
		p.a = a
	}

	p.l = len(p.a)
}

// splitByRegexp は regexp.Regexp.Split の []byte 版。
// regexp には []byte を分割するメソッドがないので、Split と同じ規則で実装している
func splitByRegexp(reg *regexp.Regexp, b []byte) [][]byte {
	if len(b) == 0 {
		return [][]byte{b}
	}

	matches := reg.FindAllIndex(b, -1)
	a := make([][]byte, 0, len(matches)+1)

	beg, end := 0, 0
	for _, match := range matches {
		end = match[0]
		if match[1] != 0 {
			a = append(a, b[beg:end])
		}
		beg = match[1]
	}

	if end != len(b) {
		a = append(a, b[beg:])
	}

	return a
}

func NewPreSplitIterator(s, sep string, re bool) *PreSplitIterator {
	a := bytes.Split([]byte(s), []byte(sep))
	if re {
		a = removeEmpty(a)
	}
	p := &PreSplitIterator{
		a:           a,
		sep:         []byte(sep),
		removeEmpty: re,
	}
	p.l = len(p.a)
	return p
}

func NewPreSplitByRegexpIterator(s string, reg *regexp.Regexp, re bool) *PreSplitIterator {
	a := splitByRegexp(reg, []byte(s))
	if re {
		a = removeEmpty(a)
	}

	p := &PreSplitIterator{
		a:           a,
		reg:         reg,
		removeEmpty: re,
	}
	p.l = len(p.a)
	return p
}
