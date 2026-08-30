# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`sel` is a Go CLI tool for column selection from text input, combining `cut(1)` functionality with `awk`-style column selection and Python-like slice notation. Built with Cobra/Viper.

## Build & Test Commands

```bash
# Build (outputs to dist/sel)
make build

# Run all tests (builds first)
make test

# Run tests only
go test -v ./...

# Run specific package tests
go test -v ./internal/column/...
go test -v ./internal/parser/...

# Run single test
go test -v ./test -run Test_E2E

# Lint
golangci-lint run
```

## Architecture

### Query Flow
1. **Parser** (`internal/parser/`) - Parses query strings into `Selector` implementations
   - Index queries: `1`, `-1`
   - Range queries: `1:10`, `1:10:2`, `-4:`
   - Switch queries (sed/awk 2addr style): `/regexp/:/regexp/`, `1:/end/`, `/start/:+3`

2. **Selectors** (`internal/column/`) - Three selector types implementing `Selector` interface:
   - `IndexSelector` - Single column by index
   - `RangeSelector` - Range with optional step (Python slice notation)
   - `SwitchSelector` - Regex-based range selection with +N/-N context support

3. **Iterators** (`internal/iterator/`) - Split into two interfaces:
   - `Source` - Supplies one `Columns` per line/record, `io.EOF` at the end. `cmd/root.go` drives this loop
     - `lineSource` - Reads lines with `bufio` and hands them to a splitting `Columns`
     - `csvSource` - Reads records with `encoding/csv` (already split)
   - `Columns` - Per-line column view (`ElementAt` / `ToArray`, both `[]byte`). The only thing `Selector` sees
     - `Iterator` - On-demand splitting
     - `RegexpIterator` - Regex-based splitting
     - `PreSplitIterator` - Pre-split all columns (for `-S` flag or CSV/TSV)

4. **Output** (`internal/output/`) - `Writer` handles delimiter joining and template-based output
   - `option.Template` (`internal/option/template.go`) parses `-t` into literal fragments. It is NOT `text/template`: `{}` is a placeholder, `{{`/`}}` are literal braces, everything else is copied as-is

### Key Design Decisions
- **1-indexed columns**: Index `0` returns the entire line (like awk's `$0`)
- **Negative indices**: `-1` is last column, `-2` is second-to-last
- **Zero-copy line reading**: columns are `[]byte` all the way from `bufio`'s read buffer to `output.Writer`, so index queries allocate nothing per line — whatever the column count — and no `unsafe` is needed. Everything a `Columns` returns points into that buffer and is only valid until the next `Source.Next()` — copy with `bytes.Clone` to keep it longer
- **Buffers are never shrunk**: `Reset` only does `[:0]` on `front`/`back`, keeping the backing array for the next line. Since `Reset` runs per line, what is retained tops out at one line's worth of columns, the same high-water-mark deal `lineSource.buf` and the `csvSource` buffers already take. Dropping the array instead makes wide lines re-allocate every line — that was #121. (`ToArray`, used by range queries, still allocates per line)
- **Zero-width separator matches are rune boundaries**: they split but never produce an empty column, so an empty separator (literal or regexp) explodes a line into runes (`bytes.Split(b, nil)` / gawk's `FS=""`); invalid UTF-8 bytes become one column each. All four `Columns` implementations must agree here — `TestEmptySeparatorAgreement` pins it
- **Lazy vs eager splitting**: Default is lazy (efficient for early columns), `-S` flag pre-splits (efficient for later columns)
- **CSV/TSV mode**: Uses `encoding/csv` for proper quote handling

### Command Flags (defined in `cmd/root.go`)
- `-d`/`-D`: Input/output delimiters
- `-g`: Use regexp for input delimiter
- `-a`: Shorthand for `-gd '\s+'`
- `-r`: Remove empty columns
- `-S`: Pre-split before selection
- `--csv`/`--tsv`: CSV/TSV parsing mode
- `-t`: Template output with `{}` placeholders (`{{`/`}}` escape a literal brace)
