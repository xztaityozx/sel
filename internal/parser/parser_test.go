package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xztaityozx/sel/internal/column"
)

func newSwitchSelector(begin, end string) column.SwitchSelector {
	s, _ := column.NewSwitchSelector(begin, end)
	return s
}

func TestParse(t *testing.T) {
	type args struct {
		queries []string
	}

	tests := []struct {
		name    string
		args    args
		want    []column.Selector
		wantErr bool
	}{
		{
			name: "1 2 3", args: args{queries: []string{"1", "2", "3"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewIndexSelector(2),
				column.NewIndexSelector(3),
			},
		},
		{
			name: "1 1:5", args: args{queries: []string{"1", "1:5"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 5, false),
			},
		},
		{
			name: "1 1:", args: args{queries: []string{"1", "1:"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 1, true),
			},
		},
		{
			name: "1 1:3:", args: args{queries: []string{"1", "1:3:"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 3, false),
			},
		},
		{
			name: "1 1:3:2", args: args{queries: []string{"1", "1:3:2"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 2, 3, false),
			},
		},
		{
			name: "1 1::2", args: args{queries: []string{"1", "1::2"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 2, 1, true),
			},
		},
		{
			name: "1 :10:", args: args{queries: []string{"1", ":10:"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 10, false),
			},
		},
		{
			name: "1 :10:4", args: args{queries: []string{"1", ":10:4"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 4, 10, false),
			},
		},
		{
			name: "1 ::", args: args{queries: []string{"1", "::"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 1, true),
			},
		},
		{
			name: "1 1::", args: args{queries: []string{"1", "1::"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				column.NewRangeSelector(1, 1, 1, true),
			},
		},
		{
			name: "1 1:/abc/", args: args{queries: []string{"1", "1:/abc/"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				newSwitchSelector("1", "/abc/"),
			},
		},
		{
			name: "1 /xyz/:/abc/", args: args{queries: []string{"1", "/xyz/:/abc/"}}, want: []column.Selector{
				column.NewIndexSelector(1),
				newSwitchSelector("/xyz/", "/abc/"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.queries)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse_RangeDirection(t *testing.T) {
	// start と stop が両方とも負でなければ、範囲の向きは行に関係なく決まる。
	// step の向きと食い違うクエリはどの行でも成立しないのでパース時に断る
	for _, q := range []string{"5:1", "1:0", "2:0", "1:5:-1", "10:12:-1"} {
		_, err := Parse([]string{q})
		require.Error(t, err, "query: %q", q)
	}

	// 負の指定は行のカラム数で解決するまで向きが決まらないので、ここでは通す
	for _, q := range []string{"-1:1:-1", "1:-5:-1", "-8:-1:2", "2:-8:-1", "5::-1", "2:2"} {
		_, err := Parse([]string{q})
		assert.NoError(t, err, "query: %q", q)
	}
}

func TestParseExclude(t *testing.T) {
	t.Run("index/rangeクエリを受け付ける", func(t *testing.T) {
		e, err := ParseExclude([]string{"1", "-1", "2:4", "2:8:2", "3:"})
		require.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("switchクエリは受け付けない", func(t *testing.T) {
		_, err := ParseExclude([]string{"/a/:/b/"})
		assert.Error(t, err)
	})

	t.Run("index 0 は受け付けない", func(t *testing.T) {
		for _, q := range []string{"0", ""} {
			_, err := ParseExclude([]string{q})
			assert.Error(t, err, "query: %q", q)
		}
	})

	t.Run("不正なクエリ", func(t *testing.T) {
		for _, q := range []string{"a", "1:2:0"} {
			_, err := ParseExclude([]string{q})
			assert.Error(t, err, "query: %q", q)
		}
	})
}
