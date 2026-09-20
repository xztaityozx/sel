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
	n := len(columns)
	start, stop := r.resolve(n)
	step := r.step

	// 開いた範囲(2: や 5::-1 など)で start が行の外側にあるのは、向きの食い違いではなく選べるカラムなし
	if r.isInfStop && ((r.step > 0 && start > stop) || (r.step < 0 && start < stop)) {
		return nil
	}

	if start == stop {
		if r.includesWholeLine() {
			// index 0 (行全体) だけを指している。単項の 0 と同じ扱い
			return w.WriteLine(columns)
		}
		if start > n || start < 1 {
			return fmt.Errorf("index %d: %w", start, iterator.ErrIndexOutOfRange)
		}
		return w.Write(columns[start-1])
	}

	if start < stop {
		if step < 0 {
			return fmt.Errorf("step must be bigger than 0(start:step:stop=%d:%d:%d)", start, step, stop)
		}
		// 書かれたままの 0 は行全体の指定なので、行内には詰めずそのまま残す
		start, stop = clampForward(start, stop, step, n, r.start != 0)
		if start > stop {
			return nil
		}
		return r.selectForward(w, columns, start, stop, step)
	}

	// start > stop
	if step > 0 {
		return fmt.Errorf("step must be less than 0(start:step:stop=%d:%d:%d)", start, step, stop)
	}
	start, stop = clampBackward(start, stop, step, n, r.stop != 0)
	if start < stop {
		return nil
	}
	return r.selectBackward(w, columns, start, stop, step)
}

// selectForward は start < stop の場合の選択処理
func (r RangeSelector) selectForward(w *output.Writer, columns [][]byte, start, stop, step int) error {
	for i := start; i <= stop; i += step {
		if i == 0 {
			// i == 0 は index 0 (行全体) の指定。空行では columns が0個になりうるが、
			// $0 は「カラムが0個」ではなく「空文字列のカラムが1個」として書く(WriteLine 参照)
			if err := w.WriteLine(columns); err != nil {
				return err
			}
		} else {
			if err := w.Write(columns[i-1]); err != nil {
				return err
			}
		}
	}
	return nil
}

// selectBackward は start > stop の場合の選択処理
func (r RangeSelector) selectBackward(w *output.Writer, columns [][]byte, start, stop, step int) error {
	for i := start; i >= stop; i += step {
		if i == 0 {
			// i == 0 は index 0 (行全体) の指定。空行では columns が0個になりうるが、
			// $0 は「カラムが0個」ではなく「空文字列のカラムが1個」として書く(WriteLine 参照)
			if err := w.WriteLine(columns); err != nil {
				return err
			}
		} else {
			if err := w.Write(columns[i-1]); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolve は行のカラム数 n に対して start/stop を実際のカラム番号に解決する。
// 行内へのクランプはここではしない。stop を n に詰めてしまうと、行より後ろを指す範囲指定
// (3カラムの行に対する 10:12 など)が start > stop になって「step の向きが違う」エラーに化ける。
// 詰めるのは範囲の向きが決まったあと (clampForward / clampBackward)
func (r RangeSelector) resolve(n int) (start, stop int) {
	start = r.start
	if start < 0 {
		start = n + start + 1
	}

	// 開いた終端は範囲の向きにある行の端。前向きなら行末、後ろ向き(5::-1 など)なら先頭カラム
	if r.isInfStop {
		if r.step < 0 {
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

// clampForward は前向きの反復範囲を行内([1, n])に詰める。詰めないと 2:100000000000 のような
// 行より後ろまで伸びた指定が行ごとに巨大なループになり、step 次第では i が溢れて負に回り込み終わらなくなる。
// 行の先頭より前を指す start は step の刻みを保ったまま最初の行内カラムまで進める。
// 詰めた結果 start > stop になったら、その行で選べるカラムは1つもない
func clampForward(start, stop, step, n int, clampStart bool) (int, int) {
	if clampStart && start < 1 {
		start = 1 + ((start-1)%step+step)%step
	}
	if stop > n {
		stop = n
	}
	return start, stop
}

// clampBackward は後ろ向きの反復範囲を行内([1, n])に詰める。理由と詰め方は clampForward と同じ
func clampBackward(start, stop, step, n int, clampStop bool) (int, int) {
	if d := -step; start > n {
		start = n - ((n-start)%d+d)%d
	}
	if clampStop && stop < 1 {
		stop = 1
	}
	return start, stop
}

// markExcluded は -x でこの範囲が指すカラムに印をつける。
// 行の外に出た添字は黙って飛ばす。解決後の範囲と step の向きが食い違うのは、負の終端が
// start より手前に落ちた場合だけ、つまりその行で選べるカラムが1つもないということなので、
// これも何もしない。行のカラム数に関係なく向きが矛盾しているクエリは parser が弾いている
func (r RangeSelector) markExcluded(mark []bool) {
	n := len(mark)
	start, stop := r.resolve(n)

	// 開いた範囲で start が行の外側にあるのも、除外するカラムなし
	if r.isInfStop && ((r.step > 0 && start > stop) || (r.step < 0 && start < stop)) {
		return
	}

	// Select と同じく、1カラムだけを指す範囲は step の向きを問わない
	if start == stop {
		markColumn(mark, start)
		return
	}

	// index 0 は NewExclusion が弾いているので、ここでは常に行内へ詰めてよい
	if start < stop {
		if r.step < 0 {
			return
		}
		start, stop = clampForward(start, stop, r.step, n, true)
		for i := start; i <= stop; i += r.step {
			markColumn(mark, i)
		}
		return
	}

	if r.step > 0 {
		return
	}
	start, stop = clampBackward(start, stop, r.step, n, true)
	for i := start; i >= stop; i += r.step {
		markColumn(mark, i)
	}
}

// includesWholeLine はクエリが index 0 (行全体) を名指ししているかどうかを返す。
// 負の値は行のカラム数で解決されるので、ここで見るのは書かれたままの 0 だけ
func (r RangeSelector) includesWholeLine() bool {
	return r.start == 0 || (!r.isInfStop && r.stop == 0)
}

// markColumn は行のカラム数に収まっている i だけに印をつける
func markColumn(mark []bool, i int) {
	if i >= 1 && i <= len(mark) {
		mark[i-1] = true
	}
}
