package iterator

import (
	"errors"
	"regexp"
	"strings"
)

// PreSplitIterator は先に全カラムへ分割してしまう Columns
type PreSplitIterator struct {
	a           []string
	sep         string
	reg         *regexp.Regexp
	l           int
	removeEmpty bool
}

func (p *PreSplitIterator) ElementAt(idx int) (string, error) {
	if p.l < idx {
		return "", errors.New(IndexOutOfRange)
	}

	if idx < 0 {
		if -p.l > idx {
			return "", errors.New(IndexOutOfRange)
		}
		return p.a[p.l+idx], nil
	}

	return p.a[idx-1], nil
}

func (p *PreSplitIterator) ToArray() []string {
	return p.a
}

func (p *PreSplitIterator) Reset(s string) {
	if p.reg == nil {
		p.resetFromArray(strings.Split(s, p.sep))
	} else {
		p.resetFromArray(p.reg.Split(s, -1))
	}
}

// resetFromArray は分割済みの配列をそのままカラム列として受け取る。
// encoding/csv のように分割済みのレコードが手に入る入力で使う
func (p *PreSplitIterator) resetFromArray(a []string) {
	if p.removeEmpty {
		p.a = removeEmpty(a)
	} else {
		p.a = a
	}

	p.l = len(p.a)
}

func NewPreSplitIterator(s, sep string, re bool) *PreSplitIterator {
	a := strings.Split(s, sep)
	if re {
		a = removeEmpty(a)
	}
	p := &PreSplitIterator{
		a:           a,
		sep:         sep,
		removeEmpty: re,
	}
	p.l = len(p.a)
	return p
}

func NewPreSplitByRegexpIterator(s string, reg *regexp.Regexp, re bool) *PreSplitIterator {
	a := reg.Split(s, -1)
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
