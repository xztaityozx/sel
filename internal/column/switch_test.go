package column

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/test"
	"regexp"
	"strings"
	"testing"
)

func TestNewSwitchSelector(t *testing.T) {
	type args struct {
		begin string
		end   string
	}
	tests := []struct {
		name    string
		args    args
		want    SwitchSelector
		wantErr bool
	}{
		{name: "Num to Num", args: args{begin: "1", end: "5"}, want: SwitchSelector{begin: address{num: 1}, end: endAddress{address: address{num: 5}}}},
		{name: "Num to /Regexp/", args: args{begin: "1", end: `/\d+/`}, want: SwitchSelector{begin: address{num: 1}, end: endAddress{address: address{regexp: regexp.MustCompile(`\d+`)}}}},
		{name: "/Regexp/ to /Regexp/", args: args{begin: `/\d+/`, end: `/\d+/`}, want: SwitchSelector{begin: address{regexp: regexp.MustCompile(`\d+`)}, end: endAddress{address: address{regexp: regexp.MustCompile(`\d+`)}}}},
		{name: "/Regexp/ after Nth", args: args{begin: `/\d+/`, end: `+5`}, want: SwitchSelector{begin: address{regexp: regexp.MustCompile(`\d+`)}, end: endAddress{address: address{num: 5}, isAroundContext: true}}},
		{name: "/Regexp/ before Nth", args: args{begin: `/\d+/`, end: `-5`}, want: SwitchSelector{begin: address{regexp: regexp.MustCompile(`\d+`)}, end: endAddress{address: address{num: -5}, isAroundContext: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSwitchSelector(tt.args.begin, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSwitchSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type testEnumerable struct {
	a []string
}

func (t *testEnumerable) ResetFromArray(_ []string) {
	panic("implement me")
}

func (t *testEnumerable) ElementAt(_ int) (string, error) {
	panic("implement me")
}

func (t *testEnumerable) Next() (item string, ok bool) {
	panic("implement me")
}

func (t *testEnumerable) Last() (item string, ok bool) {
	panic("implement me")
}

func (t *testEnumerable) Reset(_ string) {
	panic("implement me")
}

func (t *testEnumerable) ToArray() []string {
	return t.a
}

func TestSwitchSelector_Select(t *testing.T) {
	type fields struct {
		begin address
		end   endAddress
	}
	type args struct {
		w    *Writer
		iter iterator.IEnumerable
	}

	var cols []string
	for i := 0; i < 20; i++ {
		cols = append(cols, test.RandString(10))
	}

	var buf []byte
	w := bytes.NewBuffer(buf)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1 to 5",
			fields: fields{
				begin: address{num: 1},
				end:   endAddress{address: address{num: 5}},
			},
			args: args{iter: &testEnumerable{a: cols}, w: NewWriter(" ", w)},
			want: cols[0:5],
		},
		{
			name: "/a/ to /e/",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`a`)},
				end:   endAddress{address{regexp: regexp.MustCompile(`e`)}, false},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"a", "b", "c", "d", "e"},
		},
		{
			name: "/a/ after 5th",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`a`)},
				end:   endAddress{address{num: 5}, true},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"a", "b", "c", "d", "e", "3"},
		},
		{
			name: "/e/ before 5th",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`e`)},
				end:   endAddress{address{num: -5}, true},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"2", "a", "b", "c", "d", "e"},
		},
		{
			name: "/a/ before 5th",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`a`)},
				end:   endAddress{address{num: -5}, true},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"1", "2", "a"},
		},
		{
			name: "/e/ after 5th",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`e`)},
				end:   endAddress{address{num: 5}, true},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"e", "3", "4"},
		},
		{
			name: "/e/ after 0th",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`e`)},
				end:   endAddress{address{num: 0}, true},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"e"},
		},
		{
			name: "/e/ after",
			fields: fields{
				begin: address{regexp: regexp.MustCompile(`e`)},
				end:   endAddress{address{num: 0}, false},
			},
			args: args{iter: &testEnumerable{a: []string{"1", "2", "a", "b", "c", "d", "e", "3", "4"}}, w: NewWriter(" ", w)},
			want: []string{"e", "3", "4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SwitchSelector{
				begin: tt.fields.begin,
				end:   tt.fields.end,
			}
			err := s.Select(tt.args.w, tt.args.iter)
			assert.Nil(t, tt.args.w.Flush())
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, strings.Join(tt.want, " "), w.String())

			w.Reset()
		})
	}
}
