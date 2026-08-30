package option_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
)

// render は Template にカラムを流し込んだ結果を組み立てる。
// リテラルとプレースホルダの対応が期待通りかを、出力の形で確かめるためのヘルパ
func render(t *option.Template, columns []string) []byte {
	var b []byte
	for i := 0; i < t.Placeholders(); i++ {
		var column []byte
		if i < len(columns) {
			column = []byte(columns[i])
		}
		b = t.AppendColumn(b, i, column)
	}
	return t.AppendTail(b)
}

func TestParseTemplate(t *testing.T) {
	// レンダリング結果の中でカラムがどこに入ったか分かるように、カラムには C1, C2... を使う
	columns := []string{"C1", "C2", "C3"}

	tests := []struct {
		name             string
		input            string
		wantPlaceholders int
		want             string
	}{
		{name: "空文字列", input: "", wantPlaceholders: 0, want: ""},
		{name: "プレースホルダなし", input: "abc", wantPlaceholders: 0, want: "abc"},
		{name: "プレースホルダだけ", input: "{}", wantPlaceholders: 1, want: "C1"},
		{name: "前後にリテラル", input: "one: {}!", wantPlaceholders: 1, want: "one: C1!"},
		{name: "連続するプレースホルダ", input: "{}{}{}", wantPlaceholders: 3, want: "C1C2C3"},
		{name: "リテラルを挟んだ複数のプレースホルダ", input: "a{}b{}c", wantPlaceholders: 2, want: "aC1bC2c"},
		{name: "{{ はリテラルの {", input: "{{", wantPlaceholders: 0, want: "{"},
		{name: "}} はリテラルの }", input: "}}", wantPlaceholders: 0, want: "}"},
		{name: "{{}} はリテラルの {}", input: "{{}}", wantPlaceholders: 0, want: "{}"},
		{name: "text/template の記法は評価されない", input: "x{{.}}y {}", wantPlaceholders: 1, want: "x{.}y C1"},
		{name: "{{ index . 0 }} も評価されない", input: "{{ index . 0 }}", wantPlaceholders: 0, want: "{ index . 0 }"},
		{name: "単独の { はリテラル", input: "a{b", wantPlaceholders: 0, want: "a{b"},
		{name: "単独の } はリテラル", input: "a}b", wantPlaceholders: 0, want: "a}b"},
		{name: "末尾の { はリテラル", input: "a{", wantPlaceholders: 0, want: "a{"},
		{name: "末尾の } はリテラル", input: "a}", wantPlaceholders: 0, want: "a}"},
		{name: "JSON 風", input: `{"key": "{}"}`, wantPlaceholders: 1, want: `{"key": "C1"}`},
		{name: "マルチバイト混在", input: "あ{}い{}う", wantPlaceholders: 2, want: "あC1いC2う"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := option.ParseTemplate(tt.input)
			assert.Equal(t, tt.wantPlaceholders, tmpl.Placeholders())
			assert.Equal(t, tt.want, string(render(tmpl, columns)))
		})
	}
}
