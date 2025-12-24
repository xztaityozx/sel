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

3. **Iterators** (`internal/iterator/`) - Line splitting strategies via `IEnumerable` interface:
   - `Iterator` - On-demand string splitting
   - `RegexpIterator` - Regex-based splitting
   - `PreSplitIterator` - Pre-split all columns (for `-S` flag or CSV/TSV)

4. **Output** (`internal/output/`) - `Writer` handles delimiter joining and template-based output

### Key Design Decisions
- **1-indexed columns**: Index `0` returns the entire line (like awk's `$0`)
- **Negative indices**: `-1` is last column, `-2` is second-to-last
- **Lazy vs eager splitting**: Default is lazy (efficient for early columns), `-S` flag pre-splits (efficient for later columns)
- **CSV/TSV mode**: Uses `encoding/csv` for proper quote handling

### Command Flags (defined in `cmd/root.go`)
- `-d`/`-D`: Input/output delimiters
- `-g`: Use regexp for input delimiter
- `-a`: Shorthand for `-gd '\s+'`
- `-r`: Remove empty columns
- `-S`: Pre-split before selection
- `--csv`/`--tsv`: CSV/TSV parsing mode
- `-t`: Template output with `{}` placeholders
