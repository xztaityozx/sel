package parser

import (
	"github.com/xztaityozx/sel/column"
	"reflect"
	"testing"
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
