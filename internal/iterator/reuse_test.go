package iterator

import (
	"bytes"
	"regexp"
	"testing"
)

// wideLine は columns 個のカラムを空白でつないだ1行を作る。
// 幅はかつての縮小の閾値(64)を大きく超えるようにしてある
const wideColumns = 200

// shrinkThresholdWas はかつてこの容量を超えたバッファを捨てていた、その閾値。
// 伸ばしたつもりのバッファがこれを超えていないと、捨てる実装に戻しても気づけない
const shrinkThresholdWas = 64

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

// assertKeepsCapacity は grow でバッファを伸ばしてから Reset して、backing array が保たれることを見る。
// 伸びていないバッファを見ても何も確かめたことにならないので、伸びたことも一緒に確かめる
func assertKeepsCapacity(t *testing.T, name string, reset func(), grow func() error, capOf func() int) {
	t.Helper()

	reset()
	if err := grow(); err != nil {
		t.Fatal(err)
	}

	before := capOf()
	if before <= shrinkThresholdWas {
		t.Fatalf("cap(%s) = %d, want > %d: 伸びていないので Reset を見たことにならない", name, before, shrinkThresholdWas)
	}

	reset()

	if after := capOf(); after != before {
		t.Errorf("cap(%s) after Reset = %d, want %d", name, after, before)
	}
}

// TestResetKeepsCapacity は Reset が front/back の backing array を手放さないことを見る。
// 手放すと、カラム数の多い行が続くときに毎行取り直すことになって再利用の意味がなくなる。
//
// front と back は片方ずつしか伸びない。前から全カラム引くと remaining が空になって、
// 続けて負のインデックスを引いても back は伸びないので、front と back で別のイテレーターを使う
func TestResetKeepsCapacity(t *testing.T) {
	line := wideLine()

	t.Run("Iterator/front", func(t *testing.T) {
		it := NewIterator("", " ", false)
		assertKeepsCapacity(t, "front",
			func() { it.Reset(line) },
			func() error { _, err := it.ElementAt(wideColumns); return err },
			func() int { return cap(it.front) })
	})

	t.Run("Iterator/back", func(t *testing.T) {
		it := NewIterator("", " ", false)
		assertKeepsCapacity(t, "back",
			func() { it.Reset(line) },
			func() error { _, err := it.ElementAt(-wideColumns); return err },
			func() int { return cap(it.back) })
	})

	t.Run("RegexpIterator/front", func(t *testing.T) {
		it := NewRegexpIterator("", regexp.MustCompile(" "), false)
		assertKeepsCapacity(t, "front",
			func() { it.Reset(line) },
			func() error { _, err := it.ElementAt(wideColumns); return err },
			func() int { return cap(it.front) })
	})

	t.Run("RegexpIterator/back", func(t *testing.T) {
		it := NewRegexpIterator("", regexp.MustCompile(" "), false)
		assertKeepsCapacity(t, "back",
			func() { it.Reset(line) },
			func() error { _, err := it.ElementAt(-wideColumns); return err },
			func() int { return cap(it.back) })
	})
}

// TestNoAllocPerWideLine は、カラム数の多い行が続くときに行あたりのアロケーションが0であることを見る。
// 正規表現で分割するときは regexp 自体が毎回確保するので、リテラルの区切りのときだけ見られる性質
func TestNoAllocPerWideLine(t *testing.T) {
	line := wideLine()

	// TestResetKeepsCapacity と同じ理由で front と back は別のイテレーターで測る
	tests := []struct {
		name string
		idx  int
	}{
		{"front", wideColumns},
		{"back", -wideColumns},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := NewIterator("", " ", false)
			oneLine := func() {
				it.Reset(line)
				_, _ = it.ElementAt(tt.idx)
			}

			// 最初の1行はバッファを伸ばすので、測る前に流しておく
			oneLine()

			// AllocsPerRun はプロセス全体の Mallocs を見ているので、
			// テストバイナリ側のわずかな確保を拾って落ちないように少しだけ余裕をみる
			if got := testing.AllocsPerRun(100, oneLine); got > 0.5 {
				t.Errorf("allocs per line = %v, want 0", got)
			}
		})
	}
}
