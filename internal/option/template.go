package option

import "strings"

// Template は -t/--template で指定される出力テンプレート。
//
// text/template は使わない。`{}` で区切ったリテラル片をあらかじめ持っておき、
// 出力時にリテラルとカラムを交互に書き出すだけにしている。こうするとテンプレート中の
// `{{ ... }}` が Go のテンプレートとして評価されることがなく、カラム数の不一致も
// 自前のエラーにできる
type Template struct {
	// literals[i] は i 番目のプレースホルダの直前のリテラル、
	// 末尾の literals[len-1] は最後のプレースホルダの後ろのリテラル。
	// よって len(literals) == プレースホルダの数 + 1
	literals []string
}

// ParseTemplate は --template の値を Template に分解する
//
//	{}    プレースホルダ(選択されたカラムで置き換わる)
//	{{    リテラルの {
//	}}    リテラルの }
//
// 対応の取れない単独の `{` `}` はそのままリテラルとして扱う。
// `{"key": "{}"}` のような波括弧を含むテンプレートを書けるようにするため、
// エスケープ漏れをエラーにはしない
func ParseTemplate(input string) *Template {
	var literals []string
	var sb strings.Builder

	for i := 0; i < len(input); {
		switch {
		case input[i] == '{' && i+1 < len(input) && input[i+1] == '}':
			literals = append(literals, sb.String())
			sb.Reset()
			i += 2
		case input[i] == '{' && i+1 < len(input) && input[i+1] == '{':
			sb.WriteByte('{')
			i += 2
		case input[i] == '}' && i+1 < len(input) && input[i+1] == '}':
			sb.WriteByte('}')
			i += 2
		default:
			sb.WriteByte(input[i])
			i++
		}
	}

	return &Template{literals: append(literals, sb.String())}
}

// Placeholders はテンプレート中の `{}` の数を返す
func (t *Template) Placeholders() int {
	return len(t.literals) - 1
}

// AppendColumn は index 番目のプレースホルダの直前のリテラルと column を dst に足して返す
func (t *Template) AppendColumn(dst []byte, index int, column []byte) []byte {
	dst = append(dst, t.literals[index]...)
	return append(dst, column...)
}

// AppendTail は最後のプレースホルダより後ろのリテラルを dst に足して返す
func (t *Template) AppendTail(dst []byte) []byte {
	return append(dst, t.literals[len(t.literals)-1]...)
}
