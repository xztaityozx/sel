package iterator

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
)

// newEmptySepColumns は「空の区切り」を持つ splitColumns を4実装ぶん作って返す。
// newSplitColumns が返しうる組み合わせ(-S あり/なし × -g あり/なし)に対応している
func newEmptySepColumns(removeEmpty bool) map[string]splitColumns {
	return map[string]splitColumns{
		"Iterator":                 NewIterator("", "", removeEmpty),
		"PreSplitIterator":         NewPreSplitIterator("", "", removeEmpty),
		"RegexpIterator":           NewRegexpIterator("", regexp.MustCompile(``), removeEmpty),
		"PreSplitByRegexpIterator": NewPreSplitByRegexpIterator("", regexp.MustCompile(``), removeEmpty),
	}
}

// normalize は nil と長さ0のスライスの違いを吸収する。
// bytes.Split は空入力に非nilの長さ0スライスを返すが、実装によっては nil を返すため
func normalize(a []string) []string {
	if len(a) == 0 {
		return nil
	}
	return a
}

// TestEmptySeparatorAgreement は「空の区切り = ルーン分割」が4実装すべてで一致することを見る。
// 参照実装は bytes.Split(b, nil)
func TestEmptySeparatorAgreement(t *testing.T) {
	inputs := []string{"", "a", "abc", "あいう", "a\xffb", "🍣x"}

	for _, removeEmpty := range []bool{false, true} {
		for _, in := range inputs {
			want := normalize(ss(bytes.Split([]byte(in), nil)))

			for name, columns := range newEmptySepColumns(removeEmpty) {
				t.Run(fmt.Sprintf("%s/removeEmpty=%v/%q", name, removeEmpty, in), func(t *testing.T) {
					// ToArray はすべてのカラムを返す
					columns.Reset([]byte(in))
					if got := normalize(ss(columns.ToArray())); !equalStrings(got, want) {
						t.Fatalf("ToArray() = %q, want %q", got, want)
					}

					// 正のインデックスは1始まりで先頭から
					for i, w := range want {
						columns.Reset([]byte(in))
						got, err := columns.ElementAt(i + 1)
						if err != nil {
							t.Fatalf("ElementAt(%d) unexpected error: %v", i+1, err)
						}
						if string(got) != w {
							t.Errorf("ElementAt(%d) = %q, want %q", i+1, got, w)
						}
					}

					// 負のインデックスは末尾から
					for i := range want {
						idx := -(i + 1)
						w := want[len(want)-1-i]

						columns.Reset([]byte(in))
						got, err := columns.ElementAt(idx)
						if err != nil {
							t.Fatalf("ElementAt(%d) unexpected error: %v", idx, err)
						}
						if string(got) != w {
							t.Errorf("ElementAt(%d) = %q, want %q", idx, got, w)
						}
					}

					// 範囲外はエラー
					for _, idx := range []int{len(want) + 1, -(len(want) + 1)} {
						columns.Reset([]byte(in))
						if _, err := columns.ElementAt(idx); err == nil {
							t.Errorf("ElementAt(%d) expected an error, got none", idx)
						}
					}
				})
			}
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
