package output

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/option"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := &bytes.Buffer{}
	delim := "d"

	actual := NewWriter(option.Option{
		DelimiterOption: option.DelimiterOption{
			InputDelimiter: delim,
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
