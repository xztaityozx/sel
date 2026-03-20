package output

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
	"io"
	"strings"
	"testing"
	"text/template"
)

func TestNewWriter(t *testing.T) {
	w := &bytes.Buffer{}
	delim := "d"

	actual := NewWriter(option.Option{
		DelimiterOption: option.DelimiterOption{
			OutPutDelimiter: delim,
		},
	}, w, true)

	assert.NotNil(t, actual)
	assert.Equal(t, []byte(delim), actual.delimiter)
	assert.NotNil(t, actual.buf)
	assert.True(t, actual.autoFlush)
}

func TestWriter_Write(t *testing.T) {
	type fields struct {
		delimiter []byte
		buf       *bufio.Writer
	}
	type args struct {
		columns []string
	}
	cols := []string{"a", "b", "c"}
	buf := &bytes.Buffer{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "errors",
			fields: fields{
				delimiter: []byte("d"),
				buf:       bufio.NewWriter(buf),
			},
			args:    args{columns: cols},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Writer{
				delimiter: tt.fields.delimiter,
				buf:       tt.fields.buf,
			}
			if err := w.Write(tt.args.columns...); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = w.buf.Flush()

			assert.Equal(t, strings.Join(cols, "d"), buf.String())
		})
	}
}

func BenchmarkWriter_Write(b *testing.B) {
	w := NewWriter(option.Option{DelimiterOption: option.DelimiterOption{OutPutDelimiter: " "}}, io.Discard, false)
	cols := []string{"a", "b", "c", "d", "e"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Write(cols...)
		_ = w.WriteNewLine()
	}
}

func BenchmarkWriter_WriteNewLine_Template(b *testing.B) {
	// テンプレート: "{} {} {} {} {}" → "{{ index . 0 }} {{ index . 1 }} ..."
	tmplStr := ""
	for i := 0; i < 5; i++ {
		if i > 0 {
			tmplStr += " "
		}
		tmplStr += fmt.Sprintf("{{ index . %d }}", i)
	}
	tmpl := template.Must(template.New("bench").Parse(tmplStr))

	w := &Writer{
		delimiter:      []byte(" "),
		buf:            bufio.NewWriter(io.Discard),
		outputTemplate: tmpl,
		column:         []string{},
	}

	cols := []string{"a", "b", "c", "d", "e"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Write(cols...)
		_ = w.WriteNewLine()
	}
}
