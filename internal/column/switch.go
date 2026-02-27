package column

import (
	"fmt"
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/internal/output"
	"regexp"
	"strconv"
)

type address struct {
	regexp *regexp.Regexp
	num    int
}

func (a address) match(s string, index int) bool {
	if a.regexp == nil {
		return a.num-1 == index
	} else {
		return a.regexp.MatchString(s)
	}
}

// endAddress は終了アドレスを表す。終了アドレスとして+Nや-Nを指定すると、前後N個といった指定ができる
type endAddress struct {
	address
	// +N / -N といった前後を参考にするかどうか
	isAroundContext bool
}

// SwitchSelector はsedやawkにある2addrと同じ書き味でカラムを選択するやつ
type SwitchSelector struct {
	begin address
	end   endAddress
}

// between は a を丸める
// バグってるような気がしないでもない
func between(a, max, min int) int {
	if a < min {
		return 0
	} else if a > max {
		return max
	}

	return a
}

// abs は整数の絶対値を返す
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Select はクエリに従ってカラムを選択する
func (s SwitchSelector) Select(w *output.Writer, iter iterator.IEnumerable) error {
	// isAroundContextなときは、配列の最大長が必要になるので、最初に全部分割してしまう
	strings := iter.ToArray()
	maximum := len(strings)
	minimum := 0

	// スライスの初期容量を見積もる
	var estimatedCap int
	if s.end.isAroundContext {
		// AroundContext: コンテキスト幅を基準に見積もり
		contextWidth := abs(s.end.num) + 1
		if s.begin.regexp != nil {
			// 正規表現: 複数マッチの可能性があるため、大きめに確保
			// 最悪ケースでは全要素がマッチし、各マッチでcontextWidth個の要素が出力される
			estimatedCap = maximum * contextWidth
		} else {
			// インデックス指定: 1箇所のみマッチ
			estimatedCap = contextWidth
		}
	} else {
		// 通常モード: 半分程度がマッチすると仮定
		estimatedCap = maximum / 2
		if estimatedCap < 8 {
			estimatedCap = 8
		}
	}

	rt := make([]string, 0, estimatedCap)
	if s.end.isAroundContext {
		for i, v := range strings {
			if s.begin.match(v, i) {
				// マッチした位置から前後どちらかにs.end.num個
				if s.end.num < 0 {
					rt = append(rt, strings[between(i+s.end.num, maximum, minimum):between(i+1, maximum, minimum)]...)
				} else {
					rt = append(rt, strings[between(i, maximum, minimum):between(i+s.end.num+1, maximum, minimum)]...)
				}
			}
		}
	} else {
		// isAroundContextじゃないときはクエリにマッチしたとき出力するかどうかを切り替える
		// s.begin.match()でON、s.end.match()でOFFが切り替わる
		st := false
		for i, v := range strings {
			if st {
				rt = append(rt, v)
				if s.end.match(v, i) {
					st = false
				}
			} else {
				st = s.begin.match(v, i)
				if st {
					rt = append(rt, v)
				}
			}
		}
	}

	return w.Write(rt...)
}

var numberAddress, _ = regexp.Compile(`^\d+$`)
var regexpAddress, _ = regexp.Compile(`^/.+/$`)
var aroundContextAddress, _ = regexp.Compile(`^[+-]\d+$`)

func newAddress(q string) (address, error) {
	// 数値ならIndexの指定、そうでないなら正規表現としてコンパイルする
	if numberAddress.MatchString(q) {
		num, err := strconv.Atoi(q)
		return address{
			num: num,
		}, err
	} else if regexpAddress.MatchString(q) {
		r, err := regexp.Compile(q[1 : len(q)-1])
		return address{
			regexp: r,
		}, err
	}

	return address{}, fmt.Errorf("%s is not valid address", q)
}

func newEndAddress(q string) (endAddress, error) {
	// isAroundContextなクエリかどうかを最初に見る。条件としては+か-で始まる数値
	// そうでないならaddressとして解釈
	if aroundContextAddress.MatchString(q) {
		num, err := strconv.Atoi(q)
		return endAddress{
			isAroundContext: true,
			address:         address{num: num},
		}, err
	} else {
		adr, err := newAddress(q)
		return endAddress{
			address:         adr,
			isAroundContext: false,
		}, err
	}
}

func NewSwitchSelector(begin, end string) (SwitchSelector, error) {
	ba, err := newAddress(begin)
	if err != nil {
		return SwitchSelector{}, err
	}
	ea, err := newEndAddress(end)
	if err != nil {
		return SwitchSelector{}, err
	}

	return SwitchSelector{begin: ba, end: ea}, nil
}
