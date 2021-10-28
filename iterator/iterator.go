package iterator

import (
	"fmt"
	"regexp"
	"strings"
)

type IEnumerable interface {
	ElementAt(idx int) (string, error)
	Next() (item string, ok bool)
	Last() (item string, ok bool)
	ToArray() []string
	Reset(s string)
}

func removeEmpty(s []string) []string {
	a := make([]string, 0, len(s))
	for _, v := range s {
		if len(v) != 0 {
			a = append(a, v)
		}
	}
	return a
}

type Iterator struct {
	buf         map[int]string
	sep         string
	s           string
	head        int
	tail        int
	sepLen      int
	removeEmpty bool
}

var IndexOutOfRange = "index out of range"

func (i *Iterator) String() string {
	return fmt.Sprintf("{\n\tsep: '%s',\n\tsepLen: %d,\n\thead: %d,\n\ttail: %d\n\ts: '%s',\n\tbuf: %v\n}", i.sep, i.sepLen, i.head, i.tail, i.s, i.buf)
}

//Reset はこのイテレーターをリセットする
func (i *Iterator) Reset(s string) {
	i.s = s
	i.head = 0
	i.tail = 0
}

//ElementAt は指定したインデックスの値を返す。1-indexed
func (i *Iterator) ElementAt(idx int) (string, error) {
	if idx == 0 {
		return "", fmt.Errorf(IndexOutOfRange)
	}

	if idx < 0 {
		if i.tail <= idx {
			return i.buf[idx], nil
		}

		for _, ok := i.Last(); ok && i.tail >= idx; _, ok = i.Last() {
		}

		if i.tail <= idx {
			return i.buf[idx], nil
		}

		if s, ok := i.buf[idx-i.tail+i.head+1]; ok {
			i.buf[idx] = s
			return s, nil
		}

		return "", fmt.Errorf(IndexOutOfRange)
	} else {
		if i.head >= idx {
			return i.buf[idx], nil
		}

		for _, ok := i.Next(); ok && i.head <= idx; _, ok = i.Next() {
		}

		if i.head >= idx {
			return i.buf[idx], nil
		}

		if i.head+(-i.tail) >= idx {
			if s, ok := i.buf[idx-i.head+i.tail-1]; ok {
				i.buf[idx] = s
				return s, nil
			}
		}

		return "", fmt.Errorf(IndexOutOfRange)
	}
}

//Next は先頭から次の要素を取り出す
func (i *Iterator) Next() (item string, ok bool) {
	s := i.s

	if s == "" {
		return "", false
	}

	m := strings.Index(s, i.sep)
	if m < 0 {
		i.head++
		i.buf[i.head] = s
		i.s = ""
		return s, true
	}

	a := s[:m]
	i.s = s[m+i.sepLen:]

	if i.removeEmpty && a == "" {
		return i.Next()
	}

	i.head++
	i.buf[i.head] = a

	return a, true
}

//Last は末尾から要素を取り出す
func (i *Iterator) Last() (item string, ok bool) {
	s := i.s

	if s == "" {
		return "", false
	}

	m := strings.LastIndex(s, i.sep)
	if m < 0 {
		i.tail--
		i.buf[i.tail] = s
		i.s = ""
		return s, true
	}

	a := s[m+i.sepLen:]
	i.s = s[:m]

	if i.removeEmpty && a == "" {
		return i.Last()
	}

	i.tail--
	i.buf[i.tail] = a
	return a, true
}

func (i *Iterator) ToArray() []string {
	a := make([]string, i.head)
	for k := 1; k <= i.head; k++ {
		a[k-1] = i.buf[k]
	}

	if i.s != "" {
		b := strings.Split(i.s, i.sep)
		if i.removeEmpty {
			b = removeEmpty(b)
		}
		a = append(a, b...)
	}

	for k := i.tail; k <= -1; k++ {
		a = append(a, i.buf[k])
	}

	return a
}

func NewIterator(s, sep string, removeEmpty bool) *Iterator {
	buf := make(map[int]string, 20)
	buf[0] = s
	return &Iterator{
		buf:         buf,
		sep:         sep,
		s:           s,
		head:        0,
		tail:        0,
		sepLen:      len(sep),
		removeEmpty: removeEmpty,
	}
}

type RegexpIterator struct {
	r *strings.Reader
	sep *regexp.Regexp
	s string
	head int
	tail int
	buf map[int]string
	removeEmpty                 bool
	numberOfNegativeIndexMember int
}

func (r *RegexpIterator) ElementAt(idx int) (string, error) {
	if idx == 0 {
		return "", fmt.Errorf(IndexOutOfRange)
	}

	if idx > 0 {
		if r.head >= idx {
			return r.buf[idx], nil
		}

		for _, ok := r.Next(); ok && r.head <= idx; _, ok = r.Next(){}

		if r.head >= idx {
			return r.buf[idx], nil
		}

		if r.head+(-r.tail) >= idx {
			if s, ok := r.buf[idx-r.head+r.tail-1]; ok {
				r.buf[idx] = s
				return s, nil
			}
		}

		return "", fmt.Errorf(IndexOutOfRange)
	} else {
		// 負のインデックス指定されたとき
		// rightmostなIndexの検索ができないので残りの文字列をすべて分割してしまう
		// パフォーマンス的にネック
		if item, ok := r.buf[idx]; ok {
			return item, nil
		}

		if r.s != "" {
			res := make([]string, 0)
			for m := r.sep.FindReaderIndex(r.r); m != nil; m = r.sep.FindReaderIndex(r.r) {
				s := r.s

				a := s[:m[0]]
				r.s = s[m[0]+len(s[m[0]:m[1]]):]
				r.r.Reset(r.s)

				if r.removeEmpty && a == "" {
					continue
				}

				r.tail--
				res = append(res, a)
			}

			for i, v := range res {
				r.buf[-len(res)+i] = v
			}
		}

		if item, ok := r.buf[idx]; ok {
			return item, nil
		}

		if s, ok := r.buf[idx-r.tail+r.head+1]; ok {
			r.buf[idx] = s
			return s, nil
		}

		return "", fmt.Errorf(IndexOutOfRange)
	}
}

func (r *RegexpIterator) Next() (item string, ok bool) {
	s := r.s

	if s == "" {
		return "", false
	}

	m := r.sep.FindReaderIndex(r.r)
	if m == nil {
		r.head++
		r.buf[r.head] = s
		r.s = ""
		return s, true
	}

	a := s[:m[0]]
	r.s = s[m[0]+len(s[m[0]:m[1]]):]
	r.r.Reset(r.s)

	if r.removeEmpty && a == "" {
		return r.Next()
	}

	r.head++
	r.buf[r.head] = a

	return a, true
}

func (r *RegexpIterator) Last() (item string, ok bool) {
	panic("not implement Last() for RegexpIterator")
}

func (r *RegexpIterator) ToArray() []string {
	for _, ok := r.Next(); ok; _, ok = r.Next() {}

	a := make([]string, r.head+(-r.tail))
	for i:=1;i<=r.head;i++ {
		a[i-1] = r.buf[i]
	}

	for i:=-1;i>=r.tail;i-- {
		a[r.head-i+1] = r.buf[i]
	}

	return a
}

func (r *RegexpIterator) Reset(s string) {
	r.s = s
	r.r.Reset(s)
	r.head = 0
	r.tail = 0
}

func NewRegexpIterator(s string, sep *regexp.Regexp, re bool) *RegexpIterator {
	return &RegexpIterator{
		r:   strings.NewReader(s),
		sep: sep,
		s:   s,
		head: 0,
		tail: 0,
		buf: make(map[int]string, 20),
		removeEmpty: re,
	}
}

