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
			a:           bs("a", "b", "c", "d"),
			reg:         regexp.MustCompile(`\d+`),
			l:           4,
			removeEmpty: false,
		}},
		{name: "", args: args{s: "a11b22c33d", reg: regexp.MustCompile(`\d`), re: true}, want: &PreSplitIterator{
			a:           bs("a", "b", "c", "d"),
			reg:         regexp.MustCompile(`\d`),
			l:           4,
			removeEmpty: true,
		}},
		{name: "", args: args{s: "a11b22c33d", reg: regexp.MustCompile(`\d`), re: false}, want: &PreSplitIterator{
			a:           bs("a", "", "b", "", "c", "", "d"),
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
			a:           bs("a", "b", "c", "d"),
			sep:         []byte(" "),
			reg:         nil,
			l:           4,
			removeEmpty: false,
		}},
		{name: "split by space(remove-empty)", args: args{s: "a b   c d", sep: " ", re: true}, want: &PreSplitIterator{
			a:           bs("a", "b", "c", "d"),
			sep:         []byte(" "),
			reg:         nil,
			l:           4,
			removeEmpty: true,
		}},
		{name: "split by space(remove-empty)", args: args{s: "a b   c d", sep: " ", re: false}, want: &PreSplitIterator{
			a:           bs("a", "b", "", "", "c", "d"),
			sep:         []byte(" "),
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
		{name: "contained", fields: fields{a: []string{"a", "b", "c"}}, want: []string{"a", "b", "c"}},
		{name: "empty", fields: fields{a: nil}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           bs(tt.fields.a...),
				sep:         []byte(tt.fields.sep),
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			if got := ss(p.ToArray()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreSplitIterator_Reset(t *testing.T) {
	type fields struct {
		a           []string
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
		{name: "", fields: fields{a: []string{"1", "2"}, sep: " ", reg: nil, l: 2, removeEmpty: false}, args: args{s: "a b c d"}},
		{name: "", fields: fields{a: []string{"1", "2"}, sep: "", reg: regexp.MustCompile(`\d+`), l: 2, removeEmpty: false}, args: args{s: "a11b22c33d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           bs(tt.fields.a...),
				sep:         []byte(tt.fields.sep),
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}

			p.Reset([]byte(tt.args.s))

			as := assert.New(t)
			as.Equal(4, p.l)
			as.Equal(bs("a", "b", "c", "d"), p.a)
			if p.reg == nil {
				as.Nil(p.reg)
				as.Equal([]byte(" "), p.sep)
			} else {
				as.Equal(regexp.MustCompile(`\d+`), p.reg)
			}
			as.False(p.removeEmpty)
		})
	}
}

func TestPreSplitIterator_ElementAt(t *testing.T) {
	type fields struct {
		a           []string
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
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 1}, want: "1", wantErr: false},
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 4}, want: "4", wantErr: false},
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: 5}, want: "", wantErr: true},
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -5}, want: "", wantErr: true},
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -4}, want: "1", wantErr: false},
		{name: "", fields: fields{a: []string{"1", "2", "3", "4"}, sep: " ", reg: nil, l: 4, removeEmpty: false}, args: args{idx: -1}, want: "4", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PreSplitIterator{
				a:           bs(tt.fields.a...),
				sep:         []byte(tt.fields.sep),
				reg:         tt.fields.reg,
				l:           tt.fields.l,
				removeEmpty: tt.fields.removeEmpty,
			}
			got, err := p.ElementAt(tt.args.idx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElementAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("ElementAt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
