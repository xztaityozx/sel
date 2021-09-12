package column

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
	"time"
)
import "github.com/xztaityozx/sel/test_util"

func TestNewIndexSelector(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	expect := rand.Int()
	actual := NewIndexSelector(expect)
	assert.Equal(t, expect, actual.index)
}

func TestIndexSelector_Select(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		var cols []string
		for k := 0; k < 10; k++ {
			cols = append(cols, test_util.RandString(10))
		}

		is := IndexSelector{index: rand.Int() % 10}

		actual, err := is.Select(cols)
		assert.Nil(t, err)
		assert.NotNil(t, actual)
		if is.index == 0 {
			assert.Equal(t, cols, actual)
		} else {
			assert.Equal(t, 1, len(actual))
			assert.Equal(t, cols[is.index], actual[0])
		}
	}
}

func TestNewRangeSelector(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	start := rand.Int()
	step := rand.Int()
	stop := rand.Int()

	actual := NewRangeSelector(start, step, stop, true)

	assert.Equal(t, start, actual.start)
	assert.Equal(t, stop, actual.stop)
	assert.Equal(t, step, actual.step)
	assert.True(t, actual.isInfStop)
}

func TestRangeSelector_Select(t *testing.T) {
	var cols []string
	for i := 0; i < 20; i++ {
		cols = append(cols, test_util.RandString(10))
	}

	expectFactory := func(list []int) []string {
		var rt []string
		for _, v := range list {
			rt = append(rt, cols[v])
		}
		return rt
	}

	t.Run("OK", func(t *testing.T) {
		dataset := []struct {
			start   int
			step    int
			stop    int
			expects []int
		}{
			{start: 1, step: 1, stop: 5, expects: []int{1, 2, 3, 4, 5}},
			{start: 5, step: -1, stop: 1, expects: []int{5, 4, 3, 2, 1}},
			{start: 1, step: 1, stop: 1, expects: []int{0}},
			{start: -1, step: -1, stop: -5, expects: []int{19, 18, 17, 16, 15}},
			{start: 1, step: 2, stop: 10, expects: []int{1, 3, 5, 7, 9}},
			{start: -1, step: -2, stop: -10, expects: []int{19, 17, 15, 13, 11}},
		}

		for _, v := range dataset {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			expect := expectFactory(v.expects)
			actual, err := rs.Select(cols)
			assert.Nil(t, err)
			assert.Equal(t, expect, actual)
		}
	})

	t.Run("NG", func(t *testing.T) {
		dataset := []struct {
			start int
			step  int
			stop  int
		}{
			{start: 0, step: -1, stop: 5},
			{start: 5, step: 1, stop: 0},
			{start: 1000, step: 1, stop: 1000},
		}

		for _, v := range dataset {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			actual, err := rs.Select(cols)
			assert.NotNil(t, err)
			assert.Nil(t, actual)
		}
	})

	t.Run("Inf", func(t *testing.T) {
		rs := NewRangeSelector(0, 1, 0, true)
		actual, err := rs.Select(cols)
		assert.Nil(t, err)
		assert.Equal(t, cols, actual)
	})
}

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
