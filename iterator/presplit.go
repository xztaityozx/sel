package iterator

import (
	"fmt"
	"regexp"
	"strings"
)

type PreSplitIterator struct {
	a           []string
	head        int
	tail        int
	sep         string
	reg         *regexp.Regexp
	l           int
	removeEmpty bool
}

func (p *PreSplitIterator) ElementAt(idx int) (string, error) {
	if p.l < idx {
		return "", fmt.Errorf(IndexOutOfRange)
	}

	if idx < 0 {
		if -p.l > idx {
			return "", fmt.Errorf(IndexOutOfRange)
		}
		return p.a[p.l+idx], nil
	}

	return p.a[idx-1], nil
}

func (p *PreSplitIterator) Next() (item string, ok bool) {
	if p.l <= p.head {
		return "", false
	}

	if -p.l >= p.tail {
		return "", false
	}

	a := p.a[p.head]
	p.head++
	return a, true
}

func (p *PreSplitIterator) Last() (item string, ok bool) {
	if -p.l >= p.tail {
		return "", false
	}

	if p.l <= p.head {
		return "", false
	}

	a := p.a[p.l+p.tail-1]
	p.tail--
	return a, true
}

func (p *PreSplitIterator) ToArray() []string {
	return p.a
}

func (p *PreSplitIterator) Reset(s string) {
	if p.reg == nil {
		p.a = strings.Split(s, p.sep)
	} else {
		p.a = p.reg.Split(s, -1)
	}

	if p.removeEmpty {
		p.a = removeEmpty(p.a)
	}

	p.tail = 0
	p.head = 0
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
		head:        0,
		tail:        0,
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
		head:        0,
		tail:        0,
		removeEmpty: re,
	}
	p.l = len(p.a)
	return p
}
