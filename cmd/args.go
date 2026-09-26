package cmd

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// normalizeArgs はコマンドライン引数をフラグとクエリに仕分けて、flags... -- queries... の形に並べ替える。
//
// cobra/pflag にそのまま渡すと -1 や -2:-1 のような負の index クエリが shorthand フラグとして解釈されてしまうので、
// Execute() の前にクエリを "--" の後ろに逃がしておく。判定ルールは次のとおり:
//
//   - "-" の直後が数字のトークンはクエリ。sel の shorthand に数字は1つもないので、フラグと衝突しない
//   - それ以外の "-" で始まるトークンはフラグ。値を取るフラグ（-d , や --files file）は次のトークンも消費する。
//     -d, や --delimiter=, のような結合形は1トークンで完結する
//   - "--" は取り除き、その後ろも同じルールで仕分けを続ける。有効なクエリが "-" + 数字以外で始まることはないので、
//     sel -- -1 -f ./file のように "--" の後ろに書いたフラグもフラグとして扱える
//   - それ以外はクエリ
//
// 先頭がサブコマンド（completion や cobra が補完に使う __complete など）のときは、仕分けずにそのまま返す
func normalizeArgs(cmd *cobra.Command, args []string) []string {
	if len(args) == 0 || isSubCommand(cmd, args[0]) {
		return args
	}

	fs := cmd.Flags()
	flags := make([]string, 0, len(args))
	queries := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			continue
		case len(arg) < 2 || arg[0] != '-' || isDigit(arg[1]):
			queries = append(queries, arg)
		case strings.HasPrefix(arg, "--"):
			flags = append(flags, arg)
			name, _, hasValue := strings.Cut(arg[2:], "=")
			if !hasValue && takesValue(fs.Lookup(name)) && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		default:
			flags = append(flags, arg)
			if shorthandsTakeNextArg(fs, arg[1:]) && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		}
	}

	if len(queries) == 0 {
		return flags
	}
	return slices.Concat(flags, []string{"--"}, queries)
}

// shorthandsTakeNextArg は -rd のような shorthand の束が、次のトークンを値として消費するかを返す。
// 値を取るフラグが束の末尾にあるときだけ消費する（-d, や -rd, のように値が結合されていれば消費しない）
func shorthandsTakeNextArg(fs *pflag.FlagSet, shorthands string) bool {
	for i := range len(shorthands) {
		flag := fs.ShorthandLookup(shorthands[i : i+1])
		if flag == nil {
			// 未知の shorthand（と、cobra が Execute 時に足す -h / -v）は値を取らないものとみなす。
			// 未知のものは pflag がエラーにしてくれる
			continue
		}
		if takesValue(flag) {
			return i == len(shorthands)-1
		}
	}
	return false
}

// takesValue はフラグが値を取るかを返す。bool フラグのように値を省略できるものは NoOptDefVal が設定されている。
// 未知のフラグ（と、cobra が Execute 時に足す --help / --version）は値を取らないものとみなす
func takesValue(flag *pflag.Flag) bool {
	return flag != nil && flag.NoOptDefVal == ""
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

// isSubCommand は name がサブコマンド（シェル補完用の隠しコマンドを含む）の名前かを返す
func isSubCommand(cmd *cobra.Command, name string) bool {
	switch name {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return true
	}
	for _, c := range cmd.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return true
		}
	}
	return false
}
