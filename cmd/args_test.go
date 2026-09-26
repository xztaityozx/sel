package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_normalizeArgs(t *testing.T) {
	testcases := []struct {
		name     string
		args     []string
		expected []string
	}{
		{name: "no args", args: []string{}, expected: []string{}},
		{name: "positive index only", args: []string{"1", "2"}, expected: []string{"--", "1", "2"}},
		{name: "negative index", args: []string{"-1"}, expected: []string{"--", "-1"}},
		{name: "negative range", args: []string{"-2:-1"}, expected: []string{"--", "-2:-1"}},
		{name: "negative range with step", args: []string{"-10::2"}, expected: []string{"--", "-10::2"}},
		{name: "range starts with colon", args: []string{":-1"}, expected: []string{"--", ":-1"}},
		{name: "switch query", args: []string{"/^a/:/^b/"}, expected: []string{"--", "/^a/:/^b/"}},
		{name: "lone dash is a query", args: []string{"-"}, expected: []string{"--", "-"}},
		{name: "bool shorthand", args: []string{"-S", "-1"}, expected: []string{"-S", "--", "-1"}},
		{name: "bool shorthands cluster", args: []string{"-rS", "-1"}, expected: []string{"-rS", "--", "-1"}},
		{name: "value shorthand consumes next arg", args: []string{"-d", ",", "-1"}, expected: []string{"-d", ",", "--", "-1"}},
		{name: "value shorthand consumes negative number as value", args: []string{"-d", "-1", "-1"}, expected: []string{"-d", "-1", "--", "-1"}},
		{name: "value shorthand with joined value", args: []string{"-d,", "-1"}, expected: []string{"-d,", "--", "-1"}},
		{name: "value shorthand with equal", args: []string{"-d=,", "-1"}, expected: []string{"-d=,", "--", "-1"}},
		{name: "cluster ends with value shorthand", args: []string{"-rd", ",", "-1"}, expected: []string{"-rd", ",", "--", "-1"}},
		{name: "cluster with joined value", args: []string{"-rd,", "-1"}, expected: []string{"-rd,", "--", "-1"}},
		{name: "cluster value shorthand in the middle", args: []string{"-dr", "-1"}, expected: []string{"-dr", "--", "-1"}},
		{name: "long bool flag", args: []string{"--csv", "-1"}, expected: []string{"--csv", "--", "-1"}},
		{name: "long value flag consumes next arg", args: []string{"--input-delimiter", ",", "-1"}, expected: []string{"--input-delimiter", ",", "--", "-1"}},
		{name: "long value flag with equal", args: []string{"--input-delimiter=,", "-1"}, expected: []string{"--input-delimiter=,", "--", "-1"}},
		{name: "long bool flag with equal", args: []string{"--csv=true", "-1"}, expected: []string{"--csv=true", "--", "-1"}},
		{name: "flags after queries", args: []string{"-1", "-f", "./file", "2"}, expected: []string{"-f", "./file", "--", "-1", "2"}},
		{name: "flags after double dash", args: []string{"--", "-1", "-f", "./file"}, expected: []string{"-f", "./file", "--", "-1"}},
		{name: "double dash in the middle", args: []string{"-d,", "1", "2", "--", "-1", "-2"}, expected: []string{"-d,", "--", "1", "2", "-1", "-2"}},
		{name: "value flag at the end", args: []string{"-1", "-d"}, expected: []string{"-d", "--", "-1"}},
		{name: "help flag", args: []string{"--help"}, expected: []string{"--help"}},
		{name: "help shorthand", args: []string{"-h"}, expected: []string{"-h"}},
		{name: "version flag", args: []string{"--version"}, expected: []string{"--version"}},
		{name: "unknown shorthand is left to pflag", args: []string{"-x", "-1"}, expected: []string{"-x", "--", "-1"}},
		{name: "unknown long flag is left to pflag", args: []string{"--unknown", "-1"}, expected: []string{"--unknown", "--", "-1"}},
		{name: "template value that looks like a flag", args: []string{"-t", "-{}-", "-1"}, expected: []string{"-t", "-{}-", "--", "-1"}},
		{name: "completion subcommand is left as is", args: []string{"completion", "bash"}, expected: []string{"completion", "bash"}},
		{name: "help subcommand is left as is", args: []string{"help", "completion"}, expected: []string{"help", "completion"}},
		{name: "shell completion request is left as is", args: []string{"__complete", "-d", ""}, expected: []string{"__complete", "-d", ""}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, normalizeArgs(rootCmd, tc.args))
		})
	}
}
