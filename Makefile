.DEFAULT_GOAL := all
DIST_DIR := dist
EXECUTABLE_FILE := $(DIST_DIR)/sel

# clean と build/test は同じ dist/ を触るので、並列実行されると rm -r dist が
# go build と競合する。この Makefile は並列化しても得るものがない
# (go build も go test も内部で並列化する) ので、全体を直列に倒しておく。
# .WAIT でも同じことができるが、あちらは GNU Make 4.4 以降が必要で
# macOS 標準の make (3.81) では動かない
.NOTPARALLEL:

.PHONY: all
all: clean build test

.PHONY: clean
clean:
	test -e "$(DIST_DIR)" && rm -r $(DIST_DIR) || true

# find は dist/ と、.git や .claude/worktrees のようなドットディレクトリを除外する。
# './.*' は '.' 自身にはマッチしないので、リポジトリ全体が prune されることはない
$(EXECUTABLE_FILE): go.mod go.sum $(shell find . -path './.*' -prune -o -path './$(DIST_DIR)' -prune -o -name '*.go' -print)
	@mkdir -p $(@D)
	@go build -o $(DIST_DIR) -ldflags="-s -w -X github.com/xztaityozx/sel/cmd.Version=develop($(shell git rev-parse HEAD))"

.PHONY: build
build: $(EXECUTABLE_FILE)

.PHONY: test
test: clean build
	@go test -v ./...

.PHONY: ci-lint
ci-lint:
	mise exec -- golangci-lint run
