GO=go
GOFMT=gofmt
DELETE=rm
BINARY=term-poc
BUILD_BINARY=bin/$(BINARY)
# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
# current git version short-hash
VER = $(shell git rev-parse --short HEAD)

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

build: clean fmt
	$(GO) build -o $(BUILD_BINARY) -v main.go

.PHONY: clean
clean:
	$(DELETE) -rf bin/
	$(GO) clean -cache

.PHONY: fmt
fmt:
	$(GOFMT) -l -w $(SRC)

release/%: fmt
	@echo "build GOOS: $(subst release/,,$@) & GOARCH: amd64"
	GOOS=$(subst release/,,$@) GOARCH=amd64 $(GO) build -o bin/$(subst release/,,$@)/$(BINARY) -v main.go

.PHONY: test
test: build
	$(GO) test -race ./...

.PHONY: test
coverage: build
	$(GO) test -covermode=atomic --race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out