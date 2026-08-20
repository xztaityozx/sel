package iterator

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

// resetByteSlices はスライスを長さ0にリセットする。
// 容量が shrinkThreshold を超えている場合は nil を返し、backing array を GC 可能にする。
// それ以外の場合は [:0] で容量を維持して再利用する。
const shrinkThreshold = 64

func resetByteSlices(s [][]byte) [][]byte {
	if cap(s) > shrinkThreshold {
		return nil
	}
	return s[:0]
}

func removeEmpty(s [][]byte) [][]byte {
	a := make([][]byte, 0, len(s))
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
	front [][]byte
	// 後方から分割した結果 (index 0 = 最後の要素 = -1)
	back [][]byte
	// 未分割の残り
	remaining []byte
	// 区切り文字
	sep []byte
	// 区切り文字列の長さ
	sepLen int
	// 長さ0な文字列を要素に含めるかどうか
	removeEmpty bool
	// 最終的な分割結果。ToArray したときだけ書かれる
	a [][]byte
}

var IndexOutOfRange = "index out of range"

func (i *Iterator) String() string {
	return fmt.Sprintf("{\n\tsep: '%s',\n\tsepLen: %d,\n\tfront: %s,\n\tback: %s\n\tremaining: '%s'\n}", i.sep, i.sepLen, i.front, i.back, i.remaining)
}

// Reset はこのイテレーターをリセットする
func (i *Iterator) Reset(b []byte) {
	i.remaining = b
	i.front = resetByteSlices(i.front)
	i.back = resetByteSlices(i.back)
	i.a = nil
}

// ElementAt は指定したインデックスの値を返す。1-indexed
func (i *Iterator) ElementAt(idx int) ([]byte, error) {
	if idx == 0 {
		return nil, errors.New(IndexOutOfRange)
	}

	if idx > 0 {
		// 正のインデックス: front スライスを使用
		if idx <= len(i.front) {
			return i.front[idx-1], nil
		}

		// 足りなければ next() で追加分割
		for len(i.front) < idx {
			if _, ok := i.next(); !ok {
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

		return nil, errors.New(IndexOutOfRange)
	}

	// 負のインデックス: back スライスを使用
	absIdx := -idx // -1 -> 1, -2 -> 2, ...
	if absIdx <= len(i.back) {
		return i.back[absIdx-1], nil
	}

	// 足りなければ last() で追加分割
	for len(i.back) < absIdx {
		if _, ok := i.last(); !ok {
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

	return nil, errors.New(IndexOutOfRange)
}

// next は先頭から次の要素を取り出す
func (i *Iterator) next() (item []byte, ok bool) {
	s := i.remaining

	if len(s) == 0 {
		return nil, false
	}

	m := bytes.Index(s, i.sep)
	if m < 0 {
		i.front = append(i.front, s)
		i.remaining = nil
		return s, true
	}

	a := s[:m]
	i.remaining = s[m+i.sepLen:]

	if i.removeEmpty && len(a) == 0 {
		return i.next()
	}

	i.front = append(i.front, a)
	return a, true
}

// last は末尾から要素を取り出す
func (i *Iterator) last() (item []byte, ok bool) {
	s := i.remaining

	if len(s) == 0 {
		return nil, false
	}

	m := bytes.LastIndex(s, i.sep)
	if m < 0 {
		i.back = append(i.back, s)
		i.remaining = nil
		return s, true
	}

	a := s[m+i.sepLen:]
	i.remaining = s[:m]

	if i.removeEmpty && len(a) == 0 {
		return i.last()
	}

	i.back = append(i.back, a)
	return a, true
}

func (i *Iterator) ToArray() [][]byte {
	if i.a != nil {
		return i.a
	}

	// front + remaining + back(逆順) を結合
	var a [][]byte

	// front をコピー
	if len(i.front) > 0 {
		a = make([][]byte, len(i.front), len(i.front)+len(i.back)+10)
		copy(a, i.front)
	}

	// remaining を分割して追加
	if len(i.remaining) != 0 {
		b := bytes.Split(i.remaining, i.sep)
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

func NewIterator(s, sep string, removeEmpty bool) *Iterator {
	// 初期容量を設定（平均的なカラム数を想定）
	const initialCap = 16
	return &Iterator{
		front:       make([][]byte, 0, initialCap),
		back:        make([][]byte, 0, initialCap),
		remaining:   []byte(s),
		sep:         []byte(sep),
		sepLen:      len(sep),
		removeEmpty: removeEmpty,
	}
}

// RegexpIterator は正規表現でテキストを分割するイテレーター
type RegexpIterator struct {
	// 区切りとなる正規表現
	sep *regexp.Regexp
	// 未分割の残り
	s []byte
	// 前方から分割した結果 (index 0 = 1番目の要素)
	front [][]byte
	// 後方から分割した結果 (index 0 = 最後の要素 = -1)
	back [][]byte
	// 長さ0の文字列を要素に含めるかどうか
	removeEmpty bool
	// 最終的な分割結果。ToArray したときだけ書かれる
	a [][]byte
}

func (r *RegexpIterator) ElementAt(idx int) ([]byte, error) {
	if idx == 0 {
		return nil, errors.New(IndexOutOfRange)
	}

	if idx > 0 {
		// 正のインデックス: front スライスを使用
		if idx <= len(r.front) {
			return r.front[idx-1], nil
		}

		// 足りなければ next() で追加分割
		for len(r.front) < idx {
			if _, ok := r.next(); !ok {
				break
			}
		}

		if idx <= len(r.front) {
			return r.front[idx-1], nil
		}

		// front + back の合計で到達可能かチェック
		total := len(r.front) + len(r.back)
		if idx <= total {
			backIdx := idx - len(r.front) - 1
			return r.back[len(r.back)-1-backIdx], nil
		}

		return nil, errors.New(IndexOutOfRange)
	}

	// 負のインデックス: 残りの文字列をすべて分割してから返す
	absIdx := -idx // -1 -> 1, -2 -> 2, ...
	if absIdx <= len(r.back) {
		return r.back[absIdx-1], nil
	}

	// 残りの文字列をすべて分割して back に格納
	if len(r.s) != 0 {
		res := make([][]byte, 0, 16)
		for m := r.sep.FindIndex(r.s); m != nil; m = r.sep.FindIndex(r.s) {
			a := r.s[:m[0]]
			r.s = r.s[m[1]:]

			if r.removeEmpty && len(a) == 0 {
				continue
			}

			res = append(res, a)
		}

		if len(r.s) != 0 {
			res = append(res, r.s)
			r.s = nil
		}

		// res を逆順で back に追加（back[0] = 最後の要素）
		for i := len(res) - 1; i >= 0; i-- {
			r.back = append(r.back, res[i])
		}
	}

	if absIdx <= len(r.back) {
		return r.back[absIdx-1], nil
	}

	// front + back の合計で到達可能かチェック
	total := len(r.front) + len(r.back)
	if absIdx <= total {
		frontIdx := len(r.front) - (absIdx - len(r.back))
		if frontIdx >= 0 && frontIdx < len(r.front) {
			return r.front[frontIdx], nil
		}
	}

	return nil, errors.New(IndexOutOfRange)
}

// next は先頭から次の要素を取り出す
func (r *RegexpIterator) next() (item []byte, ok bool) {
	s := r.s

	if len(s) == 0 {
		return nil, false
	}

	m := r.sep.FindIndex(s)
	if m == nil {
		r.front = append(r.front, s)
		r.s = nil
		return s, true
	}

	a := s[:m[0]]
	r.s = s[m[1]:]

	if r.removeEmpty && len(a) == 0 {
		return r.next()
	}

	r.front = append(r.front, a)

	return a, true
}

func (r *RegexpIterator) ToArray() [][]byte {
	if r.a != nil {
		return r.a
	}

	// 残りをすべて next() で分割
	for _, ok := r.next(); ok; _, ok = r.next() {
	}

	// front + back(逆順) を結合
	a := make([][]byte, 0, len(r.front)+len(r.back))
	a = append(a, r.front...)
	for j := len(r.back) - 1; j >= 0; j-- {
		a = append(a, r.back[j])
	}

	r.a = a

	return a
}

func (r *RegexpIterator) Reset(b []byte) {
	r.s = b
	r.front = resetByteSlices(r.front)
	r.back = resetByteSlices(r.back)
	r.a = nil
}

func NewRegexpIterator(s string, sep *regexp.Regexp, re bool) *RegexpIterator {
	const initialCap = 16
	return &RegexpIterator{
		sep:         sep,
		s:           []byte(s),
		front:       make([][]byte, 0, initialCap),
		back:        make([][]byte, 0, initialCap),
		removeEmpty: re,
	}
}
