package iterator

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/option"
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
			as.Equal(tt.args.s, got.s)
			as.Equal(tt.args.sep, got.sep)
			as.Equal(len(tt.args.sep), len(got.sep))
			as.Equal(0, got.head)
			as.Equal(0, got.tail)
			as.Equal(tt.args.re, got.removeEmpty)
		})
	}
}

func TestIterator_Reset(t *testing.T) {
	type fields struct {
		buf         map[int]string
		sep         string
		s           string
		head        int
		tail        int
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
		{name: "", fields: fields{buf: make(map[int]string), sep: " ", s: "before", head: 100, tail: 200, sepLen: 1, removeEmpty: false}, args: args{s: "after"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				buf:         tt.fields.buf,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}

			i.Reset(tt.args.s)

			as := assert.New(t)
			as.Equal(tt.fields.sep, i.sep)
			as.Equal(tt.fields.sepLen, i.sepLen)
			as.Equal(tt.fields.removeEmpty, i.removeEmpty)
			as.Equal(tt.args.s, i.s)
			as.Equal(0, i.head)
			as.Equal(0, i.tail)
			as.Nil(i.a)
		})
	}
}

func TestIterator_ElementAt(t *testing.T) {
	type fields struct {
		buf         map[int]string
		sep         string
		s           string
		head        int
		tail        int
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
		{name: "index out of range", wantErr: true, fields: fields{}},
		{name: "1", wantErr: false, fields: fields{buf: map[int]string{1: "abc"}, head: 1}, want: "abc", args: args{idx: 1}},
		{name: "2(index out of range)", wantErr: true, fields: fields{buf: map[int]string{1: "abc"}, head: 1, s: ""}, args: args{idx: 2}},
		{name: "-1", wantErr: false, fields: fields{buf: map[int]string{1: "abc"}, head: 1, s: ""}, want: "abc", args: args{idx: -1}},
		{name: "-1(index out of range)", wantErr: true, fields: fields{buf: map[int]string{1: "abc"}, head: 1, s: ""}, args: args{idx: -2}},
		{name: "remove-empty", wantErr: false, fields: fields{buf: map[int]string{1: "a"}, head: 1, tail: 0, s: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: 3}, want: "c"},
		{name: "remove-empty(index out of range)", wantErr: true, fields: fields{buf: map[int]string{1: "a"}, head: 1, tail: 0, s: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: 5}},
		{name: "remove-empty", wantErr: false, fields: fields{buf: map[int]string{1: "a"}, head: 1, tail: 0, s: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: -3}, want: "b"},
		{name: "remove-empty(index out of range)", wantErr: true, fields: fields{buf: map[int]string{1: "a"}, head: 1, tail: 0, s: "b    c d", sep: " ", sepLen: 1, removeEmpty: true}, args: args{idx: -5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				buf:         tt.fields.buf,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			got, err := i.ElementAt(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ElementAt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIterator_Next(t *testing.T) {
	type fields struct {
		buf         map[int]string
		sep         string
		s           string
		head        int
		tail        int
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "", wantItem: "abc", wantOk: true, fields: fields{buf: make(map[int]string), sep: " ", s: "abc def", head: 0, tail: 0, sepLen: 1, removeEmpty: false}},
		{name: "", wantItem: "def", wantOk: true, fields: fields{buf: map[int]string{1: "abc"}, sep: " ", s: "def", head: 1, tail: 0, sepLen: 1, removeEmpty: false}},
		{name: "", wantItem: "", wantOk: false, fields: fields{buf: map[int]string{1: "abc", 2: "def"}, sep: " ", s: "", head: 2, tail: 0, sepLen: 1, removeEmpty: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				buf:         tt.fields.buf,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := i.Next()
			if gotItem != tt.wantItem {
				t.Errorf("Next() gotItem = %v, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Next() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestIterator_Last(t *testing.T) {
	type fields struct {
		buf         map[int]string
		sep         string
		s           string
		head        int
		tail        int
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "", wantItem: "def", wantOk: true, fields: fields{buf: make(map[int]string), sep: " ", s: "abc def", head: 0, tail: 0, sepLen: 1, removeEmpty: false}},
		{name: "", wantItem: "abc", wantOk: true, fields: fields{buf: map[int]string{1: "abc"}, sep: " ", s: "abc", head: 0, tail: 1, sepLen: 1, removeEmpty: false}},
		{name: "", wantItem: "", wantOk: false, fields: fields{buf: map[int]string{-2: "abc", -1: "def"}, sep: " ", s: "", head: 0, tail: 2, sepLen: 1, removeEmpty: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				buf:         tt.fields.buf,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := i.Last()
			if gotItem != tt.wantItem {
				t.Errorf("Last() gotItem = %v, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Last() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestIterator_ToArray(t *testing.T) {
	type fields struct {
		buf         map[int]string
		sep         string
		s           string
		head        int
		tail        int
		sepLen      int
		removeEmpty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "", fields: fields{buf: map[int]string{1: "a", 2: "b", -1: "g"}, s: "c d e f", sep: " ", sepLen: 1, head: 2, tail: -1}, want: []string{"a", "b", "c", "d", "e", "f", "g"}},
		{name: "", fields: fields{buf: map[int]string{1: "a", 2: "b", -1: "g"}, s: "", sep: " ", sepLen: 1, head: 2, tail: -1}, want: []string{"a", "b", "g"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iterator{
				buf:         tt.fields.buf,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sepLen:      tt.fields.sepLen,
				removeEmpty: tt.fields.removeEmpty,
			}
			if got := i.ToArray(); !reflect.DeepEqual(got, tt.want) {
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

			as.Equal(tt.args.s, got.s)
			as.Equal(tt.args.sep, got.sep)
			as.NotNil(got.r)
			as.Equal(0, got.head)
			as.Equal(0, got.tail)
		})
	}
}

func TestRegexpIterator_Reset(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		head        int
		tail        int
		buf         map[int]string
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
		{name: "", fields: fields{r: strings.NewReader("a b c d e"), buf: make(map[int]string), sep: regexp.MustCompile(`\d+`), s: "a b c d", head: 100, tail: 1000, removeEmpty: false, a: []string{"a"}}, args: args{s: "1 2 3 4"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				r:           tt.fields.r,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				buf:         tt.fields.buf,
				removeEmpty: tt.fields.removeEmpty,
				a:           tt.fields.a,
			}

			r.Reset(tt.args.s)

			assert.Equal(t, 0, r.head)
			assert.Equal(t, 0, r.tail)
			assert.Equal(t, tt.args.s, r.s)
			assert.Nil(t, r.a)
		})
	}
}

func TestRegexpIterator_ToArray(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		head        int
		tail        int
		buf         map[int]string
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
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d+`), s: "a11b22c33d44e", head: 0, tail: 0, buf: make(map[int]string), removeEmpty: false, a: nil},
			want:   []string{"a", "b", "c", "d", "e"},
		},
		{
			name:   `split by \d(no remove-empty)`,
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d`), s: "a11b22c33d44e", head: 0, tail: 0, buf: make(map[int]string), removeEmpty: false, a: nil},
			want:   []string{"a", "", "b", "", "c", "", "d", "", "e"},
		},
		{
			name:   `split by \d(remove-empty)`,
			fields: fields{r: strings.NewReader("a11b22c33d44e"), sep: regexp.MustCompile(`\d`), s: "a11b22c33d44e", head: 0, tail: 0, buf: make(map[int]string), removeEmpty: true, a: nil},
			want:   []string{"a", "b", "c", "d", "e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				r:           tt.fields.r,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				buf:         tt.fields.buf,
				removeEmpty: tt.fields.removeEmpty,
				a:           tt.fields.a,
			}
			if got := r.ToArray(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegexpIterator_Last(t *testing.T) {
	r := &RegexpIterator{}

	assert.Panics(t, func() {
		r.Last()
	})
}

func TestRegexpIterator_Next(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		head        int
		tail        int
		buf         map[int]string
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
				head:        0,
				tail:        0,
				buf:         map[int]string{},
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
				head:        1,
				tail:        0,
				buf:         map[int]string{1: "a"},
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
				head:        5,
				tail:        0,
				buf:         map[int]string{1: "a", 2: "b", 3: "c", 4: "d", 5: "e"},
				removeEmpty: false,
				a:           nil,
			},
			wantItem: "",
			wantOk:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				r:           tt.fields.r,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				buf:         tt.fields.buf,
				removeEmpty: tt.fields.removeEmpty,
				a:           tt.fields.a,
			}
			gotItem, gotOk := r.Next()
			if gotItem != tt.wantItem {
				t.Errorf("Next() gotItem = %v, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Next() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestRegexpIterator_ElementAt(t *testing.T) {
	type fields struct {
		r           *strings.Reader
		sep         *regexp.Regexp
		s           string
		head        int
		tail        int
		buf         map[int]string
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
				head:        0,
				tail:        0,
				buf:         map[int]string{},
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
				head:        0,
				tail:        0,
				buf:         map[int]string{},
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
				head:        0,
				tail:        0,
				buf:         map[int]string{},
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
				head:        0,
				tail:        0,
				buf:         map[int]string{},
				removeEmpty: false,
				a:           nil,
			},
			args:    args{idx: -5},
			want:    "a",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegexpIterator{
				r:           tt.fields.r,
				sep:         tt.fields.sep,
				s:           tt.fields.s,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				buf:         tt.fields.buf,
				removeEmpty: tt.fields.removeEmpty,
				a:           tt.fields.a,
			}
			got, err := r.ElementAt(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ElementAt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIEnumerable(t *testing.T) {
  as := assert.New(t)
	type args struct {
		option option.Option
	}
	tests := []struct {
		name    string
		args    args
		wants   IEnumerable
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
      args {
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
      "to be PreSplitIterator for CSV",
      args {
        option.Option{
          Xsv: option.Xsv{Csv: true, Tsv: false},
        },
      },
      NewPreSplitIterator("", ",", false),
      false,
    },
    {
      "to be PreSplitIterator for Tsv",
      args {
        option.Option{
          Xsv: option.Xsv{Csv: false, Tsv: true},
        },
      },
      NewPreSplitIterator("", "\t", false),
      false,
    },
    {
      "to be PreSplitIterator use regexp",
      args {
        option.Option{
          DelimiterOption: option.DelimiterOption{
            UseRegexp: true,
            InputDelimiter: "a",
            SplitBefore: true,
          },
        },
      },
      NewPreSplitByRegexpIterator("", regexp.MustCompile("a"), false),
      false,
    },
    {
      "to be RegexpIterator use regexp",
      args {
        option.Option{
          DelimiterOption: option.DelimiterOption{
            UseRegexp: true,
            InputDelimiter: "a",
          },
        },
      },
      NewRegexpIterator("", regexp.MustCompile("a"), false),
      false,
    },
    {
      "fail on regexp is not invalid",
      args {
        option.Option{
          DelimiterOption: option.DelimiterOption{
            UseRegexp: true,
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
      got, err := NewIEnumerable(tt.args.option)
      if tt.wantErr {
        as.Error(err)
      } else {
        as.Equal(got, tt.wants)
      }
    })
  }
}
