package column

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSplitter(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want Splitter
	}{
		{name: "string", args: args{str: "a"}, want: Splitter{
			reg: nil,
			str: "a",
		}},
		{name: "string", args: args{str: "aaaaa"}, want: Splitter{
			reg: nil,
			str: "aaaaa",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSplitter(tt.args.str, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSplitter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSplitterRegexp(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name    string
		args    args
		want    Splitter
		wantErr bool
	}{
		{
			name:    "valid",
			args:    args{query: `\d+`},
			want:    Splitter{reg: regexp.MustCompile(`\d+`)},
			wantErr: false,
		},
		{
			name:    "invalid",
			args:    args{query: `((aaa`},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSplitterRegexp(tt.args.query, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSplitterRegexp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSplitterRegexp() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitter_Split(t *testing.T) {
	type fields struct {
		reg         *regexp.Regexp
		str         string
		removeEmpty bool
	}
	type args struct {
		line string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name:   "string split",
			fields: fields{str: " "},
			args:   args{line: "a b c d e"},
			want:   []string{"a", "b", "c", "d", "e"},
		},
		{
			name:   "regexp split",
			fields: fields{reg: regexp.MustCompile(`\d+`)},
			args:   args{line: "a0b1c2d3e44444f"},
			want:   []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			name: "string split(remove empty)",
			fields: fields{
				reg:         nil,
				str:         " ",
				removeEmpty: true,
			},
			args: args{line: "a  b  c       d     e"},
			want: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:   "regexp split(remove empty)",
			fields: fields{reg: regexp.MustCompile(`\d`), removeEmpty: true},
			args:   args{line: "a0b1c2d3e44444f"},
			want:   []string{"a", "b", "c", "d", "e", "f"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Splitter{
				reg:         tt.fields.reg,
				str:         tt.fields.str,
				removeEmpty: tt.fields.removeEmpty,
			}

			got := s.Split(tt.args.line)
			assert.Equal(t, tt.want, got)
		})
	}
}
