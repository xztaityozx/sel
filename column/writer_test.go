package column

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := &bytes.Buffer{}
	delim := "d"

	actual := NewWriter(delim, w)

	assert.NotNil(t, actual)
	assert.Equal(t, []byte(delim), actual.delimiter)
	assert.NotNil(t, actual.buf)
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
			if err := w.Write(tt.args.columns); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = w.buf.Flush()

			assert.Equal(t, strings.Join(cols, "d"), buf.String())
		})
	}
}

func TestWriter_SetAutoFlush(t *testing.T) {
	type fields struct {
		delimiter []byte
		buf       *bufio.Writer
		autoFlush bool
	}
	type args struct {
		b bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect bool
	}{
		{name: "false -> true", fields: fields{
			delimiter: nil,
			buf:       nil,
			autoFlush: false,
		}, args: args{b: true}, expect: true},
		{name: "true -> false", fields: fields{
			delimiter: nil,
			buf:       nil,
			autoFlush: true,
		}, args: args{b: false}, expect: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Writer{
				delimiter: tt.fields.delimiter,
				buf:       tt.fields.buf,
				autoFlush: tt.fields.autoFlush,
			}

			w.SetAutoFlush(tt.args.b)

			assert.Equal(t, tt.expect, w.autoFlush)
		})
	}
}
