package iterator

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
)

func TestNewSource(t *testing.T) {
	type args struct {
		option option.Option
	}
	tests := []struct {
		name        string
		args        args
		wantColumns Columns
		wantCsv     bool
		wantErr     bool
	}{
		{
			name:        "to be lineSource with Iterator",
			args:        args{option.Option{}},
			wantColumns: NewIterator("", "", false),
		},
		{
			name: "to be lineSource with RegexpIterator",
			args: args{option.Option{DelimiterOption: option.DelimiterOption{
				UseRegexp:      true,
				InputDelimiter: `\s+`,
			}}},
			wantColumns: NewRegexpIterator("", regexp.MustCompile(`\s+`), false),
		},
		{
			name:        "to be csvSource for CSV",
			args:        args{option.Option{Xsv: option.Xsv{Csv: true}}},
			wantColumns: NewPreSplitIterator("", ",", false),
			wantCsv:     true,
		},
		{
			name:        "to be csvSource for TSV",
			args:        args{option.Option{Xsv: option.Xsv{Tsv: true}}},
			wantColumns: NewPreSplitIterator("", "\t", false),
			wantCsv:     true,
		},
		{
			name: "fail on regexp is not invalid",
			args: args{option.Option{DelimiterOption: option.DelimiterOption{
				UseRegexp:      true,
				InputDelimiter: "(", // invalid regexp
			}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			got, err := NewSource(tt.args.option, strings.NewReader(""))
			if tt.wantErr {
				as.Error(err)
				return
			}

			as.NoError(err)
			if tt.wantCsv {
				src, ok := got.(*csvSource)
				as.True(ok)
				as.Equal(tt.wantColumns, src.columns)
			} else {
				src, ok := got.(*lineSource)
				as.True(ok)
				as.Equal(tt.wantColumns, src.columns)
			}
		})
	}
}

// readAll は Source を終端まで回して、各行の ToArray() を集める
func readAll(t *testing.T, src Source) [][]string {
	t.Helper()

	var rt [][]string
	for {
		columns, err := src.Next()
		if errors.Is(err, io.EOF) {
			return rt
		}
		if err != nil {
			t.Fatalf("Next() unexpected error = %v", err)
		}

		// Columns が返す []byte は次の Next() までしか有効でないので、文字列に写しておく
		rt = append(rt, ss(columns.ToArray()))
	}
}

func TestLineSource_Next(t *testing.T) {
	// bufio のバッファ(既定4096バイト)に収まらない行。何度かに分けて読むことになる
	long := strings.Repeat("x", 10000)
	longer := strings.Repeat("y", 30000)

	tests := []struct {
		name  string
		input string
		want  [][]string
	}{
		{name: "empty input", input: "", want: nil},
		{name: "lines", input: "a b\nc d\n", want: [][]string{{"a", "b"}, {"c", "d"}}},
		{name: "no trailing newline", input: "a b\nc d", want: [][]string{{"a", "b"}, {"c", "d"}}},
		// 空行も1行として供給される。カラムは0個になる
		{name: "empty line", input: "a b\n\nc d\n", want: [][]string{{"a", "b"}, nil, {"c", "d"}}},
		// バッファに収まらない長い行。前後の行が壊れないことも見る
		{
			name:  "line longer than buffer",
			input: "a b\n" + long + " tail\n" + longer + " " + long + "\nc d\n",
			want:  [][]string{{"a", "b"}, {long, "tail"}, {longer, long}, {"c", "d"}},
		},
		{
			name:  "line longer than buffer without trailing newline",
			input: "a b\n" + long + " tail",
			want:  [][]string{{"a", "b"}, {long, "tail"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := NewSource(option.Option{DelimiterOption: option.DelimiterOption{InputDelimiter: " "}}, strings.NewReader(tt.input))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, readAll(t, src))
		})
	}
}

func TestLineSource_Next_ReturnsReadError(t *testing.T) {
	wantErr := errors.New("read error")
	src := &lineSource{
		reader:  bufio.NewReader(&errReader{s: "a b\n", err: wantErr}),
		columns: NewIterator("", " ", false),
	}

	columns, err := src.Next()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, ss(columns.ToArray()))

	// 行を返しきったあとにエラーが返り、それ以降は何度呼んでも同じエラーになる
	_, err = src.Next()
	assert.ErrorIs(t, err, wantErr)
	_, err = src.Next()
	assert.ErrorIs(t, err, wantErr)
}

func TestCsvSource_Next(t *testing.T) {
	src, err := NewSource(option.Option{Xsv: option.Xsv{Csv: true}}, strings.NewReader("a,\"b,c\"\nd,e\n"))
	assert.NoError(t, err)
	assert.Equal(t, [][]string{{"a", "b,c"}, {"d", "e"}}, readAll(t, src))
}

func TestCsvSource_Next_ReturnsReadError(t *testing.T) {
	src, err := NewSource(option.Option{Xsv: option.Xsv{Csv: true}}, strings.NewReader("a,\"b\nd,e\n"))
	assert.NoError(t, err)

	_, err = src.Next()
	assert.Error(t, err)
	assert.NotErrorIs(t, err, io.EOF)
}

// errReader は s を読み切ったあとに err を返す io.Reader
type errReader struct {
	s   string
	err error
}

func (e *errReader) Read(p []byte) (int, error) {
	if e.s == "" {
		return 0, e.err
	}

	n := copy(p, e.s)
	e.s = e.s[n:]
	return n, nil
}
