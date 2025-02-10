.DEFAULTGOAL := all
DIST_DIR := dist
EXECUTABLE_FILE := $(DIST_DIR)/sel

.PHONY: all
all: clean test build

clean:
	test -e "$(DIST_DIR)" && rm -r $(DIST_DIR) || true

$(EXECUTABLE_FILE):
	@mkdir -p $(@D)
	@go build -o $(DIST_DIR) -ldflags="-s -w -X github.com/xztaityozx/sel/cmd.Version=develop($(shell git rev-parse HEAD))"

.PHONY: build
build: $(EXECUTABLE_FILE)

.PHONY: test
test: build
	@go test -v ./...
