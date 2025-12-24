package iterator

import (
  "errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xztaityozx/sel/internal/option"
)

type IEnumerable interface {
	ElementAt(idx int) (string, error)
	Next() (item string, ok bool)
	Last() (item string, ok bool)
	ToArray() []string
	Reset(s string)
	ResetFromArray(a []string)
}

// NewIEnumerable は option.Option から適切な IEnumerable を生成して返す
func NewIEnumerable(option option.Option) (IEnumerable, error) {

	if ok, comma := option.IsXsv(); ok {
		// CSV/TSVの時はencoding/csvが分割をしてくれるので、NewPreSplitIteratorを使えばよい
		return NewPreSplitIterator("", string(comma), option.RemoveEmpty), nil
	}

	if option.UseRegexp {
		r, err := regexp.Compile(option.InputDelimiter)
		if err != nil {
			return nil, err
		}

		if option.SplitBefore {
			// 事前に分割する。選択しないカラムも分割するが、後半のカラムを選択するときにはこちらが有利
			return NewPreSplitByRegexpIterator("", r, option.RemoveEmpty), nil
		} else {
			// 欲しいところまでの分割を都度行う。前の方にあるindexを選ぶほど有利
			// 負のindexを指定する場合は、末尾まで分割してから返すような実装なので、実行速度が低下してしまうことに注意
			// もしかしたら肯定先読みとか使えば後ろからsplitできたりする？
			return NewRegexpIterator("", r, option.RemoveEmpty), nil
		}
	} else {
		if option.SplitBefore {
			return NewPreSplitIterator("", option.InputDelimiter, option.RemoveEmpty), nil
		} else {
			return NewIterator("", option.InputDelimiter, option.RemoveEmpty), nil
		}
	}
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

// Iterator は特定の文字で分割するイテレーター
type Iterator struct {
	// 前方から分割した結果 (index 0 = 1番目の要素)
	front []string
	// 後方から分割した結果 (index 0 = 最後の要素 = -1)
	back []string
	// 未分割の残り文字列
	remaining string
	// 区切り文字
	sep string
	// 区切り文字列の長さ
	sepLen int
	// 長さ0な文字列を要素に含めるかどうか
	removeEmpty bool
	// 最終的な分割結果。ToArray したときだけ書かれる
	a []string
}

var IndexOutOfRange = "index out of range"

func (i *Iterator) String() string {
	return fmt.Sprintf("{\n\tsep: '%s',\n\tsepLen: %d,\n\tfront: %v,\n\tback: %v\n\tremaining: '%s'\n}", i.sep, i.sepLen, i.front, i.back, i.remaining)
}

// Reset はこのイテレーターをリセットする
func (i *Iterator) Reset(s string) {
	i.remaining = s
	// スライスをクリアするが、容量は維持
	i.front = i.front[:0]
	i.back = i.back[:0]
	i.a = nil
}

// ElementAt は指定したインデックスの値を返す。1-indexed
func (i *Iterator) ElementAt(idx int) (string, error) {
	if idx == 0 {
		return "", errors.New(IndexOutOfRange)
	}

	if idx > 0 {
		// 正のインデックス: front スライスを使用
		if idx <= len(i.front) {
			return i.front[idx-1], nil
		}

		// 足りなければ Next() で追加分割
		for len(i.front) < idx {
			if _, ok := i.Next(); !ok {
				break
			}
		}

		if idx <= len(i.front) {
			return i.front[idx-1], nil
		}

		// front + back の合計で到達可能かチェック
		total := len(i.front) + len(i.back)
		if idx <= total {
			// back から取得（back は逆順なので変換が必要）
			backIdx := idx - len(i.front) - 1
			return i.back[len(i.back)-1-backIdx], nil
		}

		return "", errors.New(IndexOutOfRange)
	}

	// 負のインデックス: back スライスを使用
	absIdx := -idx // -1 -> 1, -2 -> 2, ...
	if absIdx <= len(i.back) {
		return i.back[absIdx-1], nil
	}

	// 足りなければ Last() で追加分割
	for len(i.back) < absIdx {
		if _, ok := i.Last(); !ok {
			break
		}
	}

	if absIdx <= len(i.back) {
		return i.back[absIdx-1], nil
	}

	// front + back の合計で到達可能かチェック
	total := len(i.front) + len(i.back)
	if absIdx <= total {
		// front から取得
		frontIdx := len(i.front) - (absIdx - len(i.back))
		if frontIdx >= 0 && frontIdx < len(i.front) {
			return i.front[frontIdx], nil
		}
	}

	return "", errors.New(IndexOutOfRange)
}

// Next は先頭から次の要素を取り出す
func (i *Iterator) Next() (item string, ok bool) {
	s := i.remaining

	if s == "" {
		return "", false
	}

	m := strings.Index(s, i.sep)
	if m < 0 {
		i.front = append(i.front, s)
		i.remaining = ""
		return s, true
	}

	a := s[:m]
	i.remaining = s[m+i.sepLen:]

	if i.removeEmpty && a == "" {
		return i.Next()
	}

	i.front = append(i.front, a)
	return a, true
}

// Last は末尾から要素を取り出す
func (i *Iterator) Last() (item string, ok bool) {
	s := i.remaining

	if s == "" {
		return "", false
	}

	m := strings.LastIndex(s, i.sep)
	if m < 0 {
		i.back = append(i.back, s)
		i.remaining = ""
		return s, true
	}

	a := s[m+i.sepLen:]
	i.remaining = s[:m]

	if i.removeEmpty && a == "" {
		return i.Last()
	}

	i.back = append(i.back, a)
	return a, true
}

func (i *Iterator) ToArray() []string {
	if i.a != nil {
		return i.a
	}

	// front + remaining + back(逆順) を結合
	var a []string

	// front をコピー
	if len(i.front) > 0 {
		a = make([]string, len(i.front), len(i.front)+len(i.back)+10)
		copy(a, i.front)
	}

	// remaining を分割して追加
	if i.remaining != "" {
		b := strings.Split(i.remaining, i.sep)
		if i.removeEmpty {
			b = removeEmpty(b)
		}
		a = append(a, b...)
	}

	// back を逆順で追加
	for j := len(i.back) - 1; j >= 0; j-- {
		a = append(a, i.back[j])
	}

	i.a = a
	return a
}

func (i *Iterator) ResetFromArray(_ []string) {
	panic("not impl")
}

func NewIterator(s, sep string, removeEmpty bool) *Iterator {
	// 初期容量を設定（平均的なカラム数を想定）
	const initialCap = 16
	return &Iterator{
		front:       make([]string, 0, initialCap),
		back:        make([]string, 0, initialCap),
		remaining:   s,
		sep:         sep,
		sepLen:      len(sep),
		removeEmpty: removeEmpty,
	}
}

// RegexpIterator は正規表現でテキストを分割するイテレーター
type RegexpIterator struct {
	// 入力ソース
	r *strings.Reader
	// 区切りとなる正規表現
	sep *regexp.Regexp
	// オリジナルの文字列
	s string
	// 切り出した先頭
	head int
	// 切り出した末尾
	tail int
	// 切り出し結果を保持しておくmap
	buf map[int]string
	// 長さ0の文字列を要素に含めるかどうか
	removeEmpty bool
	// 最終的な分割結果。ToArray したときだけ書かれる
	a []string
}

func (r *RegexpIterator) ElementAt(idx int) (string, error) {
	if idx == 0 {
		return "", errors.New(IndexOutOfRange)
	}

	if idx > 0 {
		if r.head >= idx {
			return r.buf[idx], nil
		}

		for _, ok := r.Next(); ok && r.head <= idx; _, ok = r.Next() {
		}

		if r.head >= idx {
			return r.buf[idx], nil
		}

		if r.head+(-r.tail) >= idx {
			if s, ok := r.buf[idx-r.head+r.tail-1]; ok {
				r.buf[idx] = s
				return s, nil
			}
		}

		return "", errors.New(IndexOutOfRange)
	} else {
		// 負のインデックス指定されたとき
		// rightmostなIndexの検索ができないので残りの文字列をすべて分割してしまう
		// パフォーマンス的にネック
		if r.tail <= idx {
			return r.buf[idx], nil
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

			if r.s != "" {
				res = append(res, r.s)
				r.s = ""
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

		return "", errors.New(IndexOutOfRange)
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
	if r.a != nil {
		return r.a
	}

	for _, ok := r.Next(); ok; _, ok = r.Next() {
	}

	a := make([]string, r.head+(-r.tail))
	for i := 1; i <= r.head; i++ {
		a[i-1] = r.buf[i]
	}

	for i := -1; i >= r.tail; i-- {
		a[r.head-i+1] = r.buf[i]
	}

	r.a = a

	return a
}

func (r *RegexpIterator) Reset(s string) {
	r.s = s
	r.r.Reset(s)
	r.head = 0
	r.tail = 0
	r.a = nil
}

func (r *RegexpIterator) ResetFromArray(_ []string) {
	panic("not impl")
}

func NewRegexpIterator(s string, sep *regexp.Regexp, re bool) *RegexpIterator {
	return &RegexpIterator{
		r:           strings.NewReader(s),
		sep:         sep,
		s:           s,
		head:        0,
		tail:        0,
		buf:         make(map[int]string, 20),
		removeEmpty: re,
	}
}
