package iterator

import (
	"bytes"
	"regexp"
	"testing"
)

// wideLine は columns 個のカラムを空白でつないだ1行を作る。
// 幅はかつての縮小の閾値(64)を大きく超えるようにしてある
const wideColumns = 200

func wideLine() []byte {
	var b bytes.Buffer
	for i := 0; i < wideColumns; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteByte('x')
	}
	return b.Bytes()
}

// TestResetKeepsCapacity は Reset が front/back の backing array を手放さないことを見る。
// 手放すと、カラム数の多い行が続くときに毎行取り直すことになって再利用の意味がなくなる
func TestResetKeepsCapacity(t *testing.T) {
	line := wideLine()

	t.Run("Iterator", func(t *testing.T) {
		it := NewIterator("", " ", false)
		it.Reset(line)
		if _, err := it.ElementAt(wideColumns); err != nil {
			t.Fatal(err)
		}
		if _, err := it.ElementAt(-wideColumns); err != nil {
			t.Fatal(err)
		}

		front, back := cap(it.front), cap(it.back)
		it.Reset(line)

		if cap(it.front) != front {
			t.Errorf("cap(front) after Reset = %d, want %d", cap(it.front), front)
		}
		if cap(it.back) != back {
			t.Errorf("cap(back) after Reset = %d, want %d", cap(it.back), back)
		}
	})

	t.Run("RegexpIterator", func(t *testing.T) {
		it := NewRegexpIterator("", regexp.MustCompile(" "), false)
		it.Reset(line)
		if _, err := it.ElementAt(wideColumns); err != nil {
			t.Fatal(err)
		}

		front := cap(it.front)
		it.Reset(line)

		if cap(it.front) != front {
			t.Errorf("cap(front) after Reset = %d, want %d", cap(it.front), front)
		}
	})
}

// TestNoAllocPerWideLine は、カラム数の多い行が続くときに行あたりのアロケーションが0であることを見る。
// 正規表現で分割するときは regexp 自体が毎回確保するので、リテラルの区切りのときだけ見られる性質
func TestNoAllocPerWideLine(t *testing.T) {
	line := wideLine()
	it := NewIterator("", " ", false)

	// 最初の1行は front/back を伸ばすので、測る前に流しておく
	oneLine := func() {
		it.Reset(line)
		_, _ = it.ElementAt(wideColumns)
		_, _ = it.ElementAt(-wideColumns)
	}
	oneLine()

	if got := testing.AllocsPerRun(100, oneLine); got != 0 {
		t.Errorf("allocs per line = %v, want 0", got)
	}
}
