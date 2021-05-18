GO=go
GOFMT=gofmt
DELETE=rm
BINARY=aardy
BIN_DIR=bin
BUILD_BINARY=$(BIN_DIR)/$(BINARY)
# go source files, ignore vendor directory
SRC := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
# current git version short-hash
VER := $(shell git rev-parse --short HEAD)

info:
	@echo " target         ⾖ Description.                                    "
	@echo " ----------------------------------------------------------------- "
	@echo
	@echo " build          generate a local build ⇨ $(BUILD_BINARY)          "
	@echo " clean          clean up bin/ & go test cache                      "
	@echo " fmt            format go code files using go fmt                  "
	@echo " release/darwin generate a darwin target build                     "
	@echo " release/linux  generate a linux target build                      "
	@echo
	@echo " ------------------------------------------------------------------"

build: fmt verinfo
	$(GO) build -o $(BUILD_BINARY) -v main.go
	cp misc/**/*.sh $(BIN_DIR)/

.PHONY: clean
clean:
	$(DELETE) -rf bin/
	$(GO) clean -cache

.PHONY: fmt
fmt:
	$(GOFMT) -l -w $(SRC)

release/%: fmt verinfo
	@echo "build GOOS: $(subst release/,,$@) & GOARCH: amd64"
	GOOS=$(subst release/,,$@) GOARCH=amd64 $(GO) build -o bin/$(subst release/,,$@)/$(BINARY) -v main.go
	cp misc/**/*.sh $(BIN_DIR)/$(subst release/,,$@)

.PHONY: test
test: build
	$(GO) test -race ./...

.PHONY: test
coverage: build
	$(GO) test -covermode=atomic --race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

.PHONY: verinfo
verinfo: commit_hash.txt build_time.txt

commit_hash.txt: FORCE
	echo "$(VER)" > "$@"

build_time.txt: FORCE
	echo "$$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$@"

FORCE: