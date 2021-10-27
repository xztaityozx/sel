package column

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"testing"
	"time"
)

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
		cols = append(cols, fmt.Sprintf("%d", i))
	}

	expectFactory := func(list []int) []string {
		var rt []string
		for _, v := range list {
			rt = append(rt, cols[v])
		}
		return rt
	}

	var buf []byte
	w := bytes.NewBuffer(buf)

	t.Run("OK", func(t *testing.T) {
		dataset := []struct {
			start   int
			step    int
			stop    int
			expects []int
		}{
			{start: 1, step: 1, stop: 5, expects: []int{0, 1, 2, 3, 4}},
			{start: 5, step: -1, stop: 1, expects: []int{4, 3, 2, 1, 0}},
			{start: 1, step: 1, stop: 1, expects: []int{0}},
			{start: -1, step: -1, stop: -5, expects: []int{19, 18, 17, 16, 15}},
			{start: 1, step: 2, stop: 10, expects: []int{0, 2, 4, 6, 8}},
			{start: -1, step: -2, stop: -10, expects: []int{19, 17, 15, 13, 11}},
		}

		for _, v := range dataset {
			rs := NewRangeSelector(v.start, v.step, v.stop, false)
			expect := expectFactory(v.expects)
			writer := NewWriter(" ", w)
			err := rs.Select(writer, &testEnumerable{a: cols})
			assert.Nil(t, writer.Flush())
			assert.Nil(t, err)
			assert.Equal(t, strings.Join(expect, " "), w.String() , "start: %d, step: %d, stop: %d", v.start, v.step, v.stop)
			w.Reset()
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
			writer := NewWriter(" ", w)
			err := rs.Select(writer, &testEnumerable{a: cols})
			assert.Nil(t, writer.Flush())
			assert.NotNil(t, err)
			assert.Equal(t, 0, len(w.String()))
			w.Reset()
		}
	})

	t.Run("Inf", func(t *testing.T) {
		rs := NewRangeSelector(1, 1, 1, true)
		writer := NewWriter(" ", w)
		err := rs.Select(writer, &testEnumerable{a: cols})
		assert.Nil(t, writer.Flush())
		assert.Nil(t, err)
		assert.Equal(t, strings.Join(cols, " "), w.String())
	})
}
