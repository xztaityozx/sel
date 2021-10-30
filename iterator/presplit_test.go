package iterator

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"testing"
)

func TestNewPreSplitByRegexpIterator(t *testing.T) {
	type args struct {
		s   string
		reg *regexp.Regexp
		re  bool
	}
	tests := []struct {
		name string
		args args
		want *PreSplitIterator
	}{
		{name: "", args: args{s: "a11b22c33d", reg: regexp.MustCompile(`\d+`), re: false}, want: &PreSplitIterator{
			a:           []string{"a", "b", "c", "d"},
			head:        0,
			tail:        0,
			reg:         regexp.MustCompile(`\d+`),
			l:           4,
			removeEmpty: false,
		}},
		{name: "", args: args{s: "a11b22c33d", reg: regexp.MustCompile(`\d`), re: true}, want: &PreSplitIterator{
			a:           []string{"a", "b", "c", "d"},
			head:        0,
			tail:        0,
			reg:         regexp.MustCompile(`\d`),
			l:           4,
			removeEmpty: true,
		}},
		{name: "", args: args{s: "a11b22c33d", reg: regexp.MustCompile(`\d`), re: false}, want: &PreSplitIterator{
			a:           []string{"a", "", "b","", "c","", "d"},
			head:        0,
			tail:        0,
			reg:         regexp.MustCompile(`\d`),
			l:           7,
			removeEmpty: false,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPreSplitByRegexpIterator(tt.args.s, tt.args.reg, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPreSplitByRegexpIterator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPreSplitIterator(t *testing.T) {
	type args struct {
		s   string
		sep string
		re  bool
	}
	tests := []struct {
		name string
		args args
		want *PreSplitIterator
	}{
		{name: "split by space(no remove-empty)", args: args{s: "a b c d", sep: " ", re: false}, want: &PreSplitIterator{
			a:           []string{"a", "b", "c", "d"},
			head:        0,
			tail:        0,
			sep:         " ",
			reg:         nil,
			l:           4,
			removeEmpty: false,
		}},
		{name: "split by space(remove-empty)", args: args{s: "a b   c d", sep: " ", re: true}, want: &PreSplitIterator{
			a:           []string{"a", "b", "c", "d"},
			head:        0,
			tail:        0,
			sep:         " ",
			reg:         nil,
			l:           4,
			removeEmpty: true,
		}},
		{name: "split by space(remove-empty)", args: args{s: "a b   c d", sep: " ", re: false}, want: &PreSplitIterator{
			a:           []string{"a", "b","", "", "c", "d"},
			head:        0,
			tail:        0,
			sep:         " ",
			reg:         nil,
			l:           6,
			removeEmpty: false,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPreSplitIterator(tt.args.s, tt.args.sep, tt.args.re); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPreSplitIterator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreSplitIterator_ToArray(t *testing.T) {
	type fields struct {
		a           []string
		head        int
		tail        int
		sep         string
		reg         *regexp.Regexp
		l           int
		removeEmpty bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "contained", fields: fields{a: []string{"a","b","c"}}, want: []string{"a","b","c"}},
		{name: "empty", fields: fields{a: nil}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           tt.fields.a,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sep:         tt.fields.sep,
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			if got := p.ToArray(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreSplitIterator_Reset(t *testing.T) {
	type fields struct {
		a           []string
		head        int
		tail        int
		sep         string
		reg         *regexp.Regexp
		l           int
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
		{name: "", fields: fields{a: []string{"1", "2"}, head: 2, tail: 0, sep: " ", reg: nil, l: 2, removeEmpty: false}, args: args{s: "a b c d"}},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 2, tail: 0, sep: "", reg: regexp.MustCompile(`\d+`), l: 2, removeEmpty: false}, args: args{s: "a11b22c33d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           tt.fields.a,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sep:         tt.fields.sep,
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}

			p.Reset(tt.args.s)

			as := assert.New(t)
			as.Equal(0, p.head)
			as.Equal(0, p.tail)
			as.Equal(4, p.l)
			as.Equal([]string{"a", "b", "c", "d"}, p.a)
			if p.reg == nil {
				as.Nil(p.reg)
				as.Equal(" ", p.sep)
			} else {
				as.Equal(regexp.MustCompile(`\d+`), p.reg)
			}
			as.False(p.removeEmpty)
		})
	}
}

func TestPreSplitIterator_Last(t *testing.T) {
	type fields struct {
		a           []string
		head        int
		tail        int
		sep         string
		reg         *regexp.Regexp
		l           int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "2"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 1, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "2"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 2, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: false, wantItem: ""},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: -1, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "1"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: -2, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: false, wantItem: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           tt.fields.a,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sep:         tt.fields.sep,
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := p.Last()
			if gotItem != tt.wantItem {
				t.Errorf("Last() gotItem = %v, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Last() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestPreSplitIterator_Next(t *testing.T) {
	type fields struct {
		a           []string
		head        int
		tail        int
		sep         string
		reg         *regexp.Regexp
		l           int
		removeEmpty bool
	}
	tests := []struct {
		name     string
		fields   fields
		wantItem string
		wantOk   bool
	}{
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "1"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 1, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "2"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 2, tail: 0, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: false, wantItem: ""},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: -1, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: true, wantItem: "1"},
		{name: "", fields: fields{a: []string{"1", "2"}, head: 0, tail: -2, sep: " ", reg: nil, removeEmpty: false, l: 2}, wantOk: false, wantItem: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           tt.fields.a,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sep:         tt.fields.sep,
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			gotItem, gotOk := p.Next()
			if gotItem != tt.wantItem {
				t.Errorf("Next() gotItem = %v, want %v", gotItem, tt.wantItem)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Next() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestPreSplitIterator_ElementAt(t *testing.T) {
	type fields struct {
		a           []string
		head        int
		tail        int
		sep         string
		reg         *regexp.Regexp
		l           int
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
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 1}, want: "1", wantErr: false},
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 4}, want: "4", wantErr: false},
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 5}, want: "", wantErr: true},
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -5}, want: "", wantErr: true},
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -4}, want: "1", wantErr: false},
		{name: "", fields: fields{a:[]string{"1", "2", "3", "4"}, head: 0, tail: 0, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -1}, want: "4", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           tt.fields.a,
				head:        tt.fields.head,
				tail:        tt.fields.tail,
				sep:         tt.fields.sep,
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			got, err := p.ElementAt(tt.args.idx)
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