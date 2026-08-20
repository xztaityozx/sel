package iterator

// bs は文字列の並びを [][]byte にする。テストのテーブルを文字列のまま書くためのヘルパ
func bs(a ...string) [][]byte {
	if a == nil {
		return nil
	}

	rt := make([][]byte, len(a))
	for i, v := range a {
		rt[i] = []byte(v)
	}
	return rt
}

// ss は [][]byte を []string にする
func ss(a [][]byte) []string {
	if a == nil {
		return nil
	}

	rt := make([]string, len(a))
	for i, v := range a {
		rt[i] = string(v)
	}
	return rt
}
