package iterator

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
)

func TestNewIterator(t *testing.T) {
	type args struct {
		s   string
		sep string
		re  bool
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "1", args: args{s: "a b c d e", sep: " ", re: true}},
		{name: "1", args: args{s: "a b c d e", sep: " ", re: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewIterator(tt.args.s, tt.args.sep, tt.args.re)
			as := assert.New(t)

			as.NotNil(got)
			as.Equal(tt.args.s, string(got.remaining))
			as.Equal(tt.args.sep, string(got.sep))
			as.Equal(len(tt.args.sep), len(got.sep))
			as.Equal(0, len(got.front))
			as.Equal(0, len(got.back))
			as.Equal(tt.args.re, got.removeEmpty)
		})
	}
}

func TestIterator_Reset(t *testing.T) {
	type fields struct {
		front       []string
		back        []string
		remaining   string
		sep         string
		sepLen      int
		removeEmpty bool
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "", fields: fields{front: []string{"a", "b"}, back: []string{"z"}, sep: " ", remaining: "before", sepLen: 1, removeEmpty: false}, args: args{s: "after"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				remaining:   []byte(tt.fields.remaining),
				sep:         []byte(tt.fields.sep),
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}

			i.Reset([]byte(tt.args.s))

			as := assert.New(t)
			as.Equal(tt.fields.sep, string(i.sep))
			as.Equal(tt.fields.sepLen, i.sepLen)
			as.Equal(tt.fields.removeEmpty, i.removeEmpty)
			as.Equal(tt.args.s, string(i.remaining))
			as.Equal(0, len(i.front))
			as.Equal(0, len(i.back))
			as.Nil(i.a)
		})
	}
}

func TestIterator_ElementAt(t *testing.T) {
	type fields struct {
		front       []string
		back        []string
		remaining   string
		sep         string
		sepLen      int
		removeEmpty bool
	}
	type args struct {
		idx int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "index out of range (idx=0)", wantErr: true, fields: fields{}},
		{name: "1", wantErr: false, fields: fields{front: []string{"abc"}}, want: "abc", args: args{idx: 1}},
		{name: "2(index out of range)", wantErr: true, fields: fields{front: []string{"abc"}, remaining: ""}, args: args{idx: 2}},
		{name: "-1", wantErr: false, fields: fields{front: []string{"abc"}, remaining: ""}, want: "abc", args: args{idx: -1}},
		{name: "-1(index out of range)", wantErr: true, fields: fields{front: []string{"abc"}, remaining: ""}, args: args{idx: -2}},
		{name: "remove-empty", wantErr: false, fields: fields{front: []string{"a"}, remaining: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: 3}, want: "c"},
		{name: "remove-empty(index out of range)", wantErr: true, fields: fields{front: []string{"a"}, remaining: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: 5}},
		{name: "remove-empty negative", wantErr: false, fields: fields{front: []string{"a"}, remaining: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: -3}, want: "b"},
		{name: "remove-empty negative(index out of range)", wantErr: true, fields: fields{front: []string{"a"}, remaining: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: -5}},
		// 空の区切りは1ルーン1カラム
		{name: "empty-sep 1", fields: fields{remaining: "abc"}, args: args{idx: 1}, want: "a"},
		{name: "empty-sep 3", fields: fields{remaining: "abc"}, args: args{idx: 3}, want: "c"},
		{name: "empty-sep 4(index out of range)", wantErr: true, fields: fields{remaining: "abc"}, args: args{idx: 4}},
		{name: "empty-sep -1", fields: fields{remaining: "abc"}, args: args{idx: -1}, want: "c"},
		{name: "empty-sep -3", fields: fields{remaining: "abc"}, args: args{idx: -3}, want: "a"},
		{name: "empty-sep -4(index out of range)", wantErr: true, fields: fields{remaining: "abc"}, args: args{idx: -4}},
		{name: "empty-sep multibyte 2", fields: fields{remaining: "あいう"}, args: args{idx: 2}, want: "い"},
		{name: "empty-sep multibyte -1", fields: fields{remaining: "あいう"}, args: args{idx: -1}, want: "う"},
		// 空カラムができないので remove-empty があっても無限再帰しない
		{name: "empty-sep remove-empty 1", fields: fields{remaining: "abc", removeEmpty: true}, args: args{idx: 1}, want: "a"},
		{name: "empty-sep remove-empty -1", fields: fields{remaining: "abc", removeEmpty: true}, args: args{idx: -1}, want: "c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				remaining:   []byte(tt.fields.remaining),
				sep:         []byte(tt.fields.sep),
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			got, err := i.ElementAt(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("ElementAt() got = %s, want %v", got, tt.want)
			}
		})
	}
}

func TestIterator_next(t *testing.T) {
	type fields struct {
		front       []string
		back        []string
		remaining   string
		sep         string
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "first element", wantItem: "abc", wantOk: true, fields: fields{front: []string{}, remaining: "abc def", sep: " ", sepLen: 1, removeEmpty: false}},
		{name: "second element", wantItem: "def", wantOk: true, fields: fields{front: []string{"abc"}, remaining: "def", sep: " ", sepLen: 1, removeEmpty: false}},
		{name: "no more elements", wantItem: "", wantOk: false, fields: fields{front: []string{"abc", "def"}, remaining: "", sep: " ", sepLen: 1, removeEmpty: false}},
		// 空の区切りは1ルーンずつ返す
		{name: "empty-sep first rune", wantItem: "a", wantOk: true, fields: fields{remaining: "abc"}},
		// remove-empty つきでも無限再帰しない(クラッシュの再現)
		{name: "empty-sep first rune(remove-empty)", wantItem: "a", wantOk: true, fields: fields{remaining: "abc", removeEmpty: true}},
		{name: "empty-sep multibyte", wantItem: "あ", wantOk: true, fields: fields{remaining: "あいう"}},
		{name: "empty-sep invalid utf8 head", wantItem: "a", wantOk: true, fields: fields{remaining: "a\xffb"}},
		{name: "empty-sep invalid utf8 byte", wantItem: "\xff", wantOk: true, fields: fields{remaining: "\xffb"}},
		{name: "empty-sep no more elements", wantItem: "", wantOk: false, fields: fields{remaining: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				remaining:   []byte(tt.fields.remaining),
				sep:         []byte(tt.fields.sep),
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := i.next()
			if string(gotItem) != tt.wantItem {
				t.Errorf("next() gotItem = %s, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("next() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestIterator_last(t *testing.T) {
	type fields struct {
		front       []string
		back        []string
		remaining   string
		sep         string
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "last element", wantItem: "def", wantOk: true, fields: fields{back: []string{}, remaining: "abc def", sep: " ", sepLen: 1, removeEmpty: false}},
		{name: "second to last", wantItem: "abc", wantOk: true, fields: fields{back: []string{"def"}, remaining: "abc", sep: " ", sepLen: 1, removeEmpty: false}},
		{name: "no more elements", wantItem: "", wantOk: false, fields: fields{back: []string{"def", "abc"}, remaining: "", sep: " ", sepLen: 1, removeEmpty: false}},
		// 空の区切りは末尾から1ルーンずつ返す
		{name: "empty-sep last rune", wantItem: "c", wantOk: true, fields: fields{remaining: "abc"}},
		{name: "empty-sep last rune(remove-empty)", wantItem: "c", wantOk: true, fields: fields{remaining: "abc", removeEmpty: true}},
		{name: "empty-sep multibyte", wantItem: "う", wantOk: true, fields: fields{remaining: "あいう"}},
		{name: "empty-sep invalid utf8 tail", wantItem: "\xff", wantOk: true, fields: fields{remaining: "a\xff"}},
		{name: "empty-sep no more elements", wantItem: "", wantOk: false, fields: fields{remaining: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				remaining:   []byte(tt.fields.remaining),
				sep:         []byte(tt.fields.sep),
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := i.last()
			if string(gotItem) != tt.wantItem {
				t.Errorf("last() gotItem = %s, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("last() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestIterator_ToArray(t *testing.T) {
	type fields struct {
		front       []string
		back        []string
		remaining   string
		sep         string
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "front + remaining + back", fields: fields{front: []string{"a", "b"}, back: []string{"g"}, remaining: "c d e f", sep: " ", sepLen: 1}, want: []string{"a", "b", "c", "d", "e", "f", "g"}},
		{name: "front + back only", fields: fields{front: []string{"a", "b"}, back: []string{"g"}, remaining: "", sep: " ", sepLen: 1}, want: []string{"a", "b", "g"}},
		// 空の区切りでは bytes.Split がルーン単位に展開するので next()/last() と一致する
		{name: "empty-sep", fields: fields{remaining: "abc"}, want: []string{"a", "b", "c"}},
		{name: "empty-sep front + remaining + back", fields: fields{front: []string{"x"}, back: []string{"z"}, remaining: "ab"}, want: []string{"x", "a", "b", "z"}},
		{name: "empty-sep multibyte", fields: fields{remaining: "あい"}, want: []string{"あ", "い"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				remaining:   []byte(tt.fields.remaining),
				sep:         []byte(tt.fields.sep),
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			if got := ss(i.ToArray()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRegexpIterator(t *testing.T) {
	type args struct {
		s   string
		sep *regexp.Regexp
		re  bool
	}
	tests := []struct {
		name string
		args args
		want *RegexpIterator
	}{
		{name: "", args: args{s: "abc", sep: regexp.MustCompile(`\d+`), re: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegexpIterator(tt.args.s, tt.args.sep, tt.args.re)
			as := assert.New(t)

			as.Equal(tt.args.s, string(got.s))
			as.Equal(tt.args.sep, got.sep)
			as.Equal(0, len(got.front))
			as.Equal(0, len(got.back))
		})
	}
}

func TestRegexpIterator_Reset(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		front       []string
		back        []string
		removeEmpty bool
		a           []string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{name: "", fields: fields{r: strings.NewReader("a b c d e"), sep: regexp.MustCompile(`\d+`), s: "a b c d", front: []string{"a", "b"}, back: []string{"d", "c"}, removeEmpty: false, a: []string{"a"}}, args: args{s: "1 2 3 4"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				sep:         tt.fields.sep,
				s:           []byte(tt.fields.s),
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				removeEmpty: tt.fields.removeEmpty,
				a:           bs(tt.fields.a...),
			}

			r.Reset([]byte(tt.args.s))

			assert.Equal(t, 0, len(r.front))
			assert.Equal(t, 0, len(r.back))
			assert.Equal(t, tt.args.s, string(r.s))
			assert.Nil(t, r.a)
		})
	}
}

func TestRegexpIterator_ToArray(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		front       []string
		back        []string
		removeEmpty bool
		a           []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   `split by \d+`,
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d+`), s: "a11b22c33d44e", front: []string{}, back: []string{}, removeEmpty: false, a: nil},
			want:   []string{"a", "b", "c", "d", "e"},
		},
		{
			name:   `split by \d(no remove-empty)`,
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d`), s: "a11b22c33d44e", front: []string{}, back: []string{}, removeEmpty: false, a: nil},
			want:   []string{"a", "", "b", "", "c", "", "d", "", "e"},
		},
		{
			name:   `split by \d(remove-empty)`,
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d`), s: "a11b22c33d44e", front: []string{}, back: []string{}, removeEmpty: true, a: nil},
			want:   []string{"a", "b", "c", "d", "e"},
		},
		// 幅0のマッチはルーン境界。以前はここで無限ループしていた
		{
			name:   "空パターンはルーン単位に分割する",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "空パターン(remove-empty)",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}, removeEmpty: true},
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "空パターン(マルチバイト)",
			fields: fields{sep: regexp.MustCompile(``), s: "あい", front: []string{}, back: []string{}},
			want:   []string{"あ", "い"},
		},
		{
			// splitByRegexp(= regexp.Regexp.Split) と同じ結果になるべき
			name:   `幅0にマッチしうる x* でも -S と一致する`,
			fields: fields{sep: regexp.MustCompile(`x*`), s: "abxxcd", front: []string{}, back: []string{}},
			want:   []string{"a", "b", "c", "d"},
		},
		{
			name:   `幅0にマッチしうる \s* でも -S と一致する`,
			fields: fields{sep: regexp.MustCompile(`\s*`), s: "a b", front: []string{}, back: []string{}},
			want:   []string{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				sep:         tt.fields.sep,
				s:           []byte(tt.fields.s),
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				removeEmpty: tt.fields.removeEmpty,
				a:           bs(tt.fields.a...),
			}
			if got := ss(r.ToArray()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexpIterator_next(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		front       []string
		back        []string
		removeEmpty bool
		a           []string
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{
			name: "1番目(a)が取り出せるべき",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			wantItem: "a",
			wantOk:   true,
		},
		{
			name: "2番目(b)が取り出せるべき",
			fields: fields{
				s:           "b22c33d44e",
				r:           strings.NewReader("b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{"a"},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			wantItem: "b",
			wantOk:   true,
		},
		{
			name: "取り出せないべき",
			fields: fields{
				s:           "",
				r:           strings.NewReader(""),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{"a", "b", "c", "d", "e"},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			wantItem: "",
			wantOk:   false,
		},
		// 幅0のマッチはルーン境界として扱い、空カラムを作らずに必ず前へ進む
		{
			name:     "空パターンなら1ルーンめ(a)が取り出せるべき",
			fields:   fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			wantItem: "a",
			wantOk:   true,
		},
		{
			name:     "空パターン+remove-empty でも無限再帰しないべき",
			fields:   fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}, removeEmpty: true},
			wantItem: "a",
			wantOk:   true,
		},
		{
			name:     `x* でも1番目(a)が取り出せるべき`,
			fields:   fields{sep: regexp.MustCompile(`x*`), s: "abxxcd", front: []string{}, back: []string{}},
			wantItem: "a",
			wantOk:   true,
		},
		{
			name:     "空パターンで空文字列なら取り出せないべき",
			fields:   fields{sep: regexp.MustCompile(``), s: "", front: []string{}, back: []string{}},
			wantItem: "",
			wantOk:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				sep:         tt.fields.sep,
				s:           []byte(tt.fields.s),
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				removeEmpty: tt.fields.removeEmpty,
				a:           bs(tt.fields.a...),
			}
			gotItem, gotOk := r.next()
			if string(gotItem) != tt.wantItem {
				t.Errorf("next() gotItem = %s, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("next() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestRegexpIterator_ElementAt(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		front       []string
		back        []string
		removeEmpty bool
		a           []string
	}
	type args struct {
		idx int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1番目が取り出せるべき",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: 1},
			want:    "a",
			wantErr: false,
		},
		{
			name: "3番目が取り出せるべき",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: 3},
			want:    "c",
			wantErr: false,
		},
		{
			name: "-1番目が取り出せるべき",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: -1},
			want:    "e",
			wantErr: false,
		},
		{
			name: "-5番目が取り出せるべき",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: -5},
			want:    "a",
			wantErr: false,
		},
		{
			name: "index out of range (idx=0)",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: 0},
			want:    "",
			wantErr: true,
		},
		{
			name: "正のインデックスが範囲外",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: 100},
			want:    "",
			wantErr: true,
		},
		{
			name: "負のインデックスが範囲外",
			fields: fields{
				s:           "a11b22c33d44e",
				r:           strings.NewReader("a11b22c33d44e"),
				sep:         regexp.MustCompile(`\d+`),
				front:       []string{},
				back:        []string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: -100},
			want:    "",
			wantErr: true,
		},
		// 幅0のマッチはルーン境界。負のインデックスは以前ここで無限ループしていた
		{
			name:   "空パターンの1番目",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: 1}, want: "a",
		},
		{
			name:   "空パターンの3番目",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: 3}, want: "c",
		},
		{
			name:   "空パターンの4番目は範囲外",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: 4}, want: "", wantErr: true,
		},
		{
			name:   "空パターンの-1番目",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: -1}, want: "c",
		},
		{
			name:   "空パターンの-3番目",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: -3}, want: "a",
		},
		{
			name:   "空パターンの-4番目は範囲外",
			fields: fields{sep: regexp.MustCompile(``), s: "abc", front: []string{}, back: []string{}},
			args:   args{idx: -4}, want: "", wantErr: true,
		},
		{
			name:   `x* の-1番目`,
			fields: fields{sep: regexp.MustCompile(`x*`), s: "abxxcd", front: []string{}, back: []string{}},
			args:   args{idx: -1}, want: "d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				sep:         tt.fields.sep,
				s:           []byte(tt.fields.s),
				front:       bs(tt.fields.front...),
				back:        bs(tt.fields.back...),
				removeEmpty: tt.fields.removeEmpty,
				a:           bs(tt.fields.a...),
			}
			got, err := r.ElementAt(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("ElementAt() got = %s, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSplitColumns(t *testing.T) {
	as := assert.New(t)
	type args struct {
		option option.Option
	}
	tests := []struct {
		name    string
		args    args
		wants   splitColumns
		wantErr bool
	}{
		{
			"to be Iterator",
			args{
				option.Option{
					DelimiterOption: option.DelimiterOption{
						SplitBefore: false,
					},
				},
			},
			NewIterator("", "", false),
			false,
		},
		{
			"to be PreSplitIterator",
			args{
				option.Option{
					DelimiterOption: option.DelimiterOption{
						SplitBefore: true,
					},
				},
			},
			NewPreSplitIterator("", "", false),
			false,
		},
		{
			"to be PreSplitIterator use regexp",
			args{
				option.Option{
					DelimiterOption: option.DelimiterOption{
						UseRegexp:      true,
						InputDelimiter: "a",
						SplitBefore:    true,
					},
				},
			},
			NewPreSplitByRegexpIterator("", regexp.MustCompile("a"), false),
			false,
		},
		{
			"to be RegexpIterator use regexp",
			args{
				option.Option{
					DelimiterOption: option.DelimiterOption{
						UseRegexp:      true,
						InputDelimiter: "a",
					},
				},
			},
			NewRegexpIterator("", regexp.MustCompile("a"), false),
			false,
		},
		{
			"fail on regexp is not invalid",
			args{
				option.Option{
					DelimiterOption: option.DelimiterOption{
						UseRegexp:      true,
						InputDelimiter: "(", // invalid regexp
					},
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newSplitColumns(tt.args.option)
			if tt.wantErr {
				as.Error(err)
			} else {
				as.Equal(got, tt.wants)
			}
		})
	}
}
