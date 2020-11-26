# sel
**sel**ect columns  

![Go](https://github.com/xztaityozx/sel/workflows/Go/badge.svg)

extra _cut(1)_ command with `awk`'s column selection and slice notation.

![example](./img/example.png)

# Install
## go get
```
$ go get -u github.com/xztaityozx/sel
```

## Download binary from GitHub Releases
Download prebuild binary from [release page](https://github.com/xztaityozx/sel/releases)


## (Optional) Shell completion script
Completion script is available for bash, fish, PowerShell and zsh.

```sh
# for bash
$ eval "$(sel --completion bash -)"
# for fish
$ eval "$(sel --completion fish -)"
# for PowerShell
$ iex "$(sel --completion PowerShell -)"
# for zsh
$ eval "$(sel --completion zsh -)"
```

# Usage

```
          _ 
 ___  ___| |
/ __|/ _ \ |
\__ \  __/ |
|___/\___|_|

__sel__ect column

Usage:
  sel [flags]

Examples:
sel 1

Flags:
  -b, --backup                    make backup when enable -i/--in-place option
  -h, --help                      help for sel
  -i, --in-place                  edit files in place
  -d, --input-delimiter string    sets field delimiter(input) (default " ")
  -f, --input-files strings       input files
  -D, --output-delimiter string   sets field delimiter(output) (default " ")
  -r, --remove-empty              remove empty sequence
  -v, --version                   version for sel
```

# Features
- one-indexed
- index `0` refers to the entire line. (like `awk`)
- slice notation
- overwrite source file
