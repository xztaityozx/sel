package enumerable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRange(t *testing.T) {
	as := assert.New(t)
	data := map[string]Range{
		"1":      {start: NewPositionIndex(1), stop: NewPositionIndex(1), step: 1},
		"1:10":   {start: NewPositionIndex(1), stop: NewPositionIndex(10), step: 1},
		"1:":     {start: NewPositionIndex(1), stop: NewInfIndex(), step: 1},
		":":      {start: NewPositionIndex(1), stop: NewInfIndex(), step: 1},
		":4":     {start: NewPositionIndex(1), stop: NewPositionIndex(4), step: 1},
		"2:10:3": {start: NewPositionIndex(2), stop: NewPositionIndex(10), step: 3},
		"3:10:":  {start: NewPositionIndex(3), stop: NewPositionIndex(10), step: 1},
		"4::10":  {start: NewPositionIndex(4), stop: NewInfIndex(), step: 10},
		"5::":    {start: NewPositionIndex(5), stop: NewInfIndex(), step: 1},
		"::":     {start: NewPositionIndex(1), stop: NewInfIndex(), step: 1},
		":6:":    {start: NewPositionIndex(1), stop: NewPositionIndex(6), step: 1},
		":6:3":   {start: NewPositionIndex(1), stop: NewPositionIndex(6), step: 3},
		"::3":    {start: NewPositionIndex(1), stop: NewInfIndex(), step: 3},
		"-6":     {start: NewPositionIndex(-6), stop: NewPositionIndex(-6), step: -6},
	}

	for q, expect := range data {
		actual, err := NewRange(q)
		as.Nil(err)
		as.Equal(expect, actual, fmt.Sprintf("query: %s", q))
	}

}
