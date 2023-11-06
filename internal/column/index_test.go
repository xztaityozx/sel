package column

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/xztaityozx/sel/internal/iterator"
	"github.com/xztaityozx/sel/test"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewIndexSelectorFromString(t *testing.T) {
	type args struct {
		str string
		def int
	}
	tests := []struct {
		name    string
		args    args
		want    IndexSelector
		wantErr bool
	}{
		{name: "10", args: args{str: "10", def: 0}, want: IndexSelector{index: 10}},
		{name: "failed to parse", args: args{str: "a", def: 0}, want: IndexSelector{}, wantErr: true},
		{name: "", args: args{str: "10", def: 10}, want: IndexSelector{index: 10}},
		{name: "-1", args: args{str: "-1", def: 10}, want: IndexSelector{index: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIndexSelectorFromString(tt.args.str, tt.args.def)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIndexSelectorFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIndexSelectorFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewIndexSelector(t *testing.T) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	expect := rand.Int()
	actual := NewIndexSelector(expect)
	assert.Equal(t, expect, actual.index)
}

func TestIndexSelector_Select(t *testing.T) {

	rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		var cols []string
		for k := 0; k < 10; k++ {
			cols = append(cols, test.RandString(10))
		}

		is := IndexSelector{index: rand.Int() % 10}

		var buf []byte
		w := bytes.NewBuffer(buf)

		writer := NewWriter(" ", w)

		err := is.Select(writer, iterator.NewIterator(strings.Join(cols, " "), " ", false))

		assert.Nil(t, writer.Flush())
		assert.Nil(t, err)
		if is.index == 0 {
			assert.Equal(t, strings.Join(cols, " "), w.String())
		} else {
			assert.Equal(t, cols[is.index-1], w.String())
		}
	}
}
