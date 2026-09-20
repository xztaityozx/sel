#!/usr/bin/env bash
# base と head の sel を同じ入力・同じクエリで hyperfine にかけ、
# head の median が base の median の $THRESHOLD 倍を超えたら失敗する
set -euo pipefail

THRESHOLD="${THRESHOLD:-1.20}"
BASE_BIN="${BASE_BIN:-/tmp/sel-base}"
HEAD_BIN="${HEAD_BIN:-/tmp/sel-head}"
DATA="${DATA:-/tmp/data.txt}"
DATA_CSV="${DATA_CSV:-/tmp/data.csv}"

# 計測するクエリ: 名前<TAB>引数
CASES=$(
	cat <<-EOS
		index	1 -f "$DATA"
		index-last	-f "$DATA" -- -1
		index-late	19 -f "$DATA"
		range	3:10 -f "$DATA"
		range-step	1:20:3 -f "$DATA"
		presplit	-S 19 -f "$DATA"
		regexp-delimiter	-a 5 -f "$DATA"
		template	-t '[{}]-[{}]' 1 2 -f "$DATA"
		csv	--csv 5 -f "$DATA_CSV"
	EOS
)

failed=0
{
	echo "| case | base (ms) | head (ms) | ratio |"
	echo "| --- | --- | --- | --- |"
} >>"${GITHUB_STEP_SUMMARY:-/dev/null}"

while IFS=$'\t' read -r name args; do
	[ -n "$name" ] || continue
	hyperfine --shell=none --output=null --warmup 2 --runs 10 --style basic \
		--export-json /tmp/result.json \
		-n base "$BASE_BIN $args" \
		-n head "$HEAD_BIN $args"

	read -r base head ratio < <(jq -r '
		(.results[] | select(.command == "base") | .median) as $b
		| (.results[] | select(.command == "head") | .median) as $h
		| [$b * 1000, $h * 1000, $h / $b] | @tsv' /tmp/result.json)

	mark=""
	if [ "$(jq -n --argjson r "$ratio" --argjson t "$THRESHOLD" '$r > $t')" = "true" ]; then
		mark=" ⚠️"
		failed=1
		printf '::error::%s: head が base の %.3f 倍遅い (閾値 %s)\n' "$name" "$ratio" "$THRESHOLD"
	fi
	printf '| %s | %.1f | %.1f | %.3f%s |\n' "$name" "$base" "$head" "$ratio" "$mark" \
		>>"${GITHUB_STEP_SUMMARY:-/dev/null}"
done <<<"$CASES"

exit "$failed"
