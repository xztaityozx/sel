# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`sel` is a Go CLI tool for column selection from text input, combining `cut(1)` with `awk`-style column selection
and Python-like slice notation. Built with Cobra/Viper.

## Build & Test Commands

```bash
make build                          # Build (outputs to dist/sel)
make test                           # Build, then run all tests
go test -v ./...                    # Run tests only
go test -v ./internal/column/...    # Run one package's tests
go test -v ./test -run Test_E2E     # Run a single test
golangci-lint run                   # Lint (make ci-lint runs it via mise)
```

## Architecture

### Query Flow
1. **Parser** (`internal/parser/`) - Parses query strings into `Selector` implementations
   - Index queries: `1`, `-1`
   - Range queries: `1:10`, `1:10:2`, `-4:`
   - Switch queries (sed/awk 2addr style): `/regexp/:/regexp/`, `1:/end/`, `/start/:+3`
   - A range whose start/stop have the same sign and whose direction contradicts its step is rejected here

2. **Selectors** (`internal/column/`) - Three types implementing `Selector`:
   - `IndexSelector` - Single column by index
   - `RangeSelector` - Range with optional step (Python slice notation). `Select` and `markExcluded` walk the range via the shared `indices`
     - Only columns that exist in the line are written; only a single-index range (`4:4`) out of range is an error
     - If resolving negative bounds against the line's column count makes the direction contradict the step, that line selects nothing
   - `SwitchSelector` - Regex-based range selection with +N/-N context support

3. **Iterators** (`internal/iterator/`) - Split into two interfaces:
   - `Source` - Supplies one `Columns` per line/record, `io.EOF` at the end. `cmd/root.go` drives this loop
     - `lineSource` - Reads lines with `bufio` and hands them to a splitting `Columns`
     - `csvSource` - Reads records with `encoding/csv` (already split)
   - `Columns` - Per-line column view (`ElementAt` / `ToArray`, both `[]byte`). The only thing a `Selector` sees
     - `Iterator` - On-demand splitting
     - `RegexpIterator` - Regex-based splitting
     - `PreSplitIterator` - All columns pre-split (`-S`, CSV/TSV, and the `-x` view)

4. **Exclusion** (`internal/column/exclude.go`) - Only when `-x` is given
   - `Exclusion.Apply` wraps a line's `Columns` in a view hiding the excluded columns; selectors see the rest renumbered from 1
   - Index and range queries only (they implement `markExcluded`); switch queries and index `0` are rejected by `NewExclusion`

5. **Output** (`internal/output/`) - `Writer` handles delimiter joining and template-based output
   - `option.Template` (`internal/option/template.go`) parses `-t` into literal fragments
   - It is NOT `text/template`: `{}` is a placeholder, `{{`/`}}` are literal braces, everything else is copied as-is

### Key Design Decisions
- **1-indexed columns**: Index `0` returns the entire line (like awk's `$0`)
- **Negative indices**: `-1` is last column, `-2` is second-to-last
- **Zero-copy line reading**: columns stay `[]byte` from `bufio`'s read buffer to `output.Writer`, without `unsafe`
  - On the default lazy literal-separator path an index query allocates nothing per line, whatever the column count
  - `-S`/`--csv`/`--tsv` re-split every line, and `RegexpIterator` allocates on negative indices
  - Everything a `Columns` returns is only valid until the next `Source.Next()`; copy with `bytes.Clone` to keep it
- **Buffers are never shrunk**: `Reset` only does `[:0]` on `front`/`back`, keeping the backing array
  - Retention tops out at one line's worth of columns, like `lineSource.buf` and the `csvSource` buffers
  - Dropping the array makes wide lines re-allocate on every line. (`ToArray`, used by range queries, still allocates)
- **Zero-width separator matches are rune boundaries**: they split but never produce an empty column
  - An empty separator (literal or regexp) explodes a line into runes, like `bytes.Split(b, nil)` / gawk's `FS=""`
  - Invalid UTF-8 bytes become one column each
  - All `Columns` implementations must agree; `TestEmptySeparatorAgreement` pins it
- **Lazy vs eager splitting**: Default is lazy (efficient for early columns), `-S` pre-splits (efficient for later columns)
- **CSV/TSV mode**: Uses `encoding/csv` for proper quote handling

### Command Flags (defined in `cmd/root.go`)
`Execute()` passes args through `normalizeArgs` (`cmd/args.go`), which reorders them into `flags... -- queries...`.
A token with `-` followed by a digit is always a query (no shorthand is a digit), so negative indices need no `--`.

- `-f`: Input files
- `-d`/`-D`: Input/output delimiters
- `-g`: Use regexp for input delimiter
- `-a`: Shorthand for `-gd '\s+'`
- `-r`: Remove empty columns
- `-S`: Pre-split before selection
- `-M`: Output an empty string for out-of-range columns instead of an error
- `-E`: Fill value for out-of-range columns (implies `-M`)
- `--csv`/`--tsv`: CSV/TSV parsing mode
- `-t`: Template output with `{}` placeholders (`{{`/`}}` escape a literal brace)
- `-x`/`--exclude`: Drop columns (index or range query) before the queries run
  - The remaining columns are renumbered from 1; with no query at all, every remaining column is printed

## Git
- Do not include a `Claude-Session:` line (or any claude.ai session URL) in commit messages or pull request bodies
