package iterator

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/xztaityozx/sel/internal/sliceutil"
)

// 区切りについての約束ごと:
//
// 幅0の区切りマッチは UTF-8 のルーン境界を表す。分割はするが空カラムは作らない。
// したがって空の区切り(リテラル・正規表現とも)は行をルーン単位に分割する。
// これは bytes.Split(b, []byte{}) や gawk の FS="" と同じ規則で、
// 不正な UTF-8 バイトは1バイト1カラムになる(utf8.DecodeRune が RuneError, 1 を返すため)。
// 空入力はカラム0個。
//
// この規則は Iterator / RegexpIterator / PreSplitIterator のすべてで一致していなければならない。
// 一致していることは TestEmptySeparatorAgreement が見ている

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
	i.front = sliceutil.Reset(i.front)
	i.back = sliceutil.Reset(i.back)
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
	for {
		s := i.remaining

		if len(s) == 0 {
			return nil, false
		}

		if i.sepLen == 0 {
			// 区切りが空文字列のときは1ルーンを1カラムとして切り出す(bytes.Split と同じ規則)。
			// 不正なUTF-8でも size >= 1 が返るので必ず前に進むし、長さ0のカラムもできない
			_, size := utf8.DecodeRune(s)
			a := s[:size]
			i.remaining = s[size:]
			i.front = append(i.front, a)
			return a, true
		}

		m := bytes.Index(s, i.sep)
		if m < 0 {
			i.front = append(i.front, s)
			i.remaining = nil
			return s, true
		}

		a := s[:m]
		i.remaining = s[m+i.sepLen:]

		// sepLen >= 1 なので、読み飛ばすたびに remaining は必ず短くなる = 必ず終わる
		if i.removeEmpty && len(a) == 0 {
			continue
		}

		i.front = append(i.front, a)
		return a, true
	}
}

// last は末尾から要素を取り出す
func (i *Iterator) last() (item []byte, ok bool) {
	for {
		s := i.remaining

		if len(s) == 0 {
			return nil, false
		}

		if i.sepLen == 0 {
			// next() と同じくルーン単位。末尾から1ルーンぶん切り出す
			_, size := utf8.DecodeLastRune(s)
			a := s[len(s)-size:]
			i.remaining = s[:len(s)-size]
			i.back = append(i.back, a)
			return a, true
		}

		m := bytes.LastIndex(s, i.sep)
		if m < 0 {
			i.back = append(i.back, s)
			i.remaining = nil
			return s, true
		}

		a := s[m+i.sepLen:]
		i.remaining = s[:m]

		// next() と同様、sepLen >= 1 なので必ず終わる
		if i.removeEmpty && len(a) == 0 {
			continue
		}

		i.back = append(i.back, a)
		return a, true
	}
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

	// remaining を分割して追加。
	// sep が空のとき bytes.Split はルーン単位に展開するので、next()/last() と自動的に一致する
	// (長さ0のカラムができないので removeEmpty も no-op になる)
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
		for {
			beg, end, found := nextSepMatch(r.sep, r.s)
			if !found {
				break
			}

			a := r.s[:beg]
			r.s = r.s[end:]

			// beg == 0 のときは end > 0 なので、どちらにせよ r.s は必ず短くなる = 必ず終わる
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

// nextSepMatch は b の中の最初の「区切り」を [beg,end) で返す。
// 幅0のマッチはルーン境界を表すものとして扱い、空のカラムを作らないように、
// 先頭での幅0マッチだけ1ルーンぶん読み飛ばしてから探し直す。
// これは splitByRegexp(= regexp.Regexp.Split) が
//   - 先頭の幅0マッチを捨てる(match[1] != 0 のチェック)
//   - 直前のマッチに接する幅0マッチを FindAll が返さない
//
// のと同じ規則を、逐次分割のために書き直したもの。
// ok のとき beg > 0 || end > 0 が必ず成り立つので、呼び出し側は必ず前に進める
func nextSepMatch(reg *regexp.Regexp, b []byte) (beg, end int, ok bool) {
	m := reg.FindIndex(b)
	if m == nil {
		return 0, 0, false
	}
	if m[0] != 0 || m[1] != 0 {
		return m[0], m[1], true
	}

	// 先頭で幅0のマッチ。カラムにせず1ルーンぶん進めてから探し直す
	_, size := utf8.DecodeRune(b)
	if size == 0 { // b が空
		return 0, 0, false
	}
	m = reg.FindIndex(b[size:])
	if m == nil {
		return 0, 0, false
	}
	return size + m[0], size + m[1], true
}

// next は先頭から次の要素を取り出す
func (r *RegexpIterator) next() (item []byte, ok bool) {
	for {
		s := r.s

		if len(s) == 0 {
			return nil, false
		}

		beg, end, found := nextSepMatch(r.sep, s)
		if !found {
			r.front = append(r.front, s)
			r.s = nil
			return s, true
		}

		a := s[:beg]
		r.s = s[end:]

		// len(a) == 0 は beg == 0、つまり end > 0 なので、読み飛ばしても必ず前に進む
		if r.removeEmpty && len(a) == 0 {
			continue
		}

		r.front = append(r.front, a)

		return a, true
	}
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
	r.front = sliceutil.Reset(r.front)
	r.back = sliceutil.Reset(r.back)
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
