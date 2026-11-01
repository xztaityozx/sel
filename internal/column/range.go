package column

import (
	"fmt"

	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
)

// RangeSelector はカラムの範囲選択するやつ
type RangeSelector struct {
	start     int
	step      int
	stop      int
	isInfStop bool
}

func NewRangeSelector(start, step, stop int, isInfStop bool) RangeSelector {
	return RangeSelector{start: start, step: step, stop: stop, isInfStop: isInfStop}
}

func (r RangeSelector) Select(w *output.Writer, iter iterator.Columns) error {
	columns := iter.ToArray()
	return r.indices(len(columns), func(i int) error {
		if i == 0 {
			// index 0 (行全体)。空行では columns が0個になりうるが、
			// $0 は「カラムが0個」ではなく「空文字列のカラムが1個」として書く(WriteLine 参照)
			return w.WriteLine(columns)
		}
		return w.Write(columns[i-1])
	})
}

// markExcluded は -x でこの範囲が指すカラムに印をつける。index 0 は NewExclusion が弾いている。
// 単一の index として書いた範囲(4:4 など)が行の外を指すエラーは、-x では除外する列がないだけなので捨てる
func (r RangeSelector) markExcluded(mark []bool) {
	_ = r.indices(len(mark), func(i int) error {
		mark[i-1] = true
		return nil
	})
}

// indices は n カラムの行でこの範囲が指す添字を順に fn に渡す。渡すのは 1..n と、書かれたままの 0 (行全体) だけ。
// 行の外に出た添字は飛ばし、エラーにするのは単一の index として書いた範囲(4:4 など)が行の外を指すときだけ。
// 解決後の範囲と step の向きが食い違うのは負の指定が行のカラム数で解決された結果なので、その行で選べるカラムなしとする
// (どの行でも食い違うクエリは parser が弾いている)
func (r RangeSelector) indices(n int, fn func(i int) error) error {
	start, stop := r.resolve(n)
	step := r.step

	if start == stop {
		// 1カラムだけを指す範囲は step の向きを問わない
		if !r.isInfStop && !r.includesWholeLine() && (start < 1 || start > n) {
			return fmt.Errorf("index %d: %w", start, iterator.ErrIndexOutOfRange)
		}
		step = 1
	}
	if (start < stop && step < 0) || (start > stop && step > 0) {
		return nil
	}

	// 行内([1, n])に詰めてから回す。詰めないと 2:100000000000 のような指定が行ごとに巨大なループになる。
	// 行の外から始まる側は step の刻みを保ったまま最初の行内カラムまで進める。書かれたままの 0 は行全体なので詰めない。
	// i += step の前に終端を越えるかを見て、桁溢れで i が負に回り込み行内に戻ってくるのを防ぐ
	if step > 0 {
		if r.start != 0 && start < 1 {
			start = 1 + ((start-1)%step+step)%step
		}
		stop = min(stop, n)
		for i := start; i <= stop; i += step {
			if err := fn(i); err != nil {
				return err
			}
			if stop-i < step {
				break
			}
		}
		return nil
	}

	// -step は step == MinInt64 で桁溢れするので、step の符号のまま計算する
	if start > n {
		if m := (start - n) % step; m == 0 {
			start = n
		} else {
			start = n + m + step
		}
	}
	if (r.isInfStop || r.stop != 0) && stop < 1 {
		stop = 1
	}
	for i := start; i >= stop; i += step {
		if err := fn(i); err != nil {
			return err
		}
		if stop-i > step {
			break
		}
	}
	return nil
}

// resolve は行のカラム数 n に対して start/stop を実際のカラム番号に解決する。
// 行内へのクランプはここではしない。stop を n に詰めてしまうと、行より後ろを指す範囲指定
// (3カラムの行に対する 10:12 など)の向きが逆転する。詰めるのは向きが決まったあと(indices)
func (r RangeSelector) resolve(n int) (start, stop int) {
	start = r.start
	if start < 0 {
		start = n + start + 1
	}

	// 開いた終端は範囲の向きにある行の端。前向きなら行末、後ろ向き(5::-1 など)なら先頭カラム。
	// 0::-1 は行全体から後ろ向きなので行全体だけ(Python の a[0::-1] と同じ)
	if r.isInfStop {
		if r.step < 0 {
			if r.start == 0 {
				return start, 0
			}
			return start, 1
		}
		return start, n
	}

	stop = r.stop
	if stop < 0 {
		stop = n + stop + 1
	}

	return start, stop
}

// includesWholeLine はクエリが index 0 (行全体) を名指ししているかどうかを返す。
// 負の値は行のカラム数で解決されるので、ここで見るのは書かれたままの 0 だけ
func (r RangeSelector) includesWholeLine() bool {
	return r.start == 0 || (!r.isInfStop && r.stop == 0)
}
