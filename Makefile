GO=go
GOARCH ?= amd64
GOFMT=gofmt
DELETE=rm
BINARY=pryrite
OS_BINARY=$(BINARY)
BIN_DIR=bin
BUILD_BINARY=$(BIN_DIR)/$(BINARY)

# go source files, ignore vendor directory
SRC := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# current git version short-hash
VER := $(shell git rev-parse --short HEAD)
WATCH := (.go$$)

ifeq ($(GOOS),windows)
  OS_BINARY := $(OS_BINARY).exe
  BUILD_BINARY := $(BUILD_BINARY).exe
endif

word-dot = $(word $2,$(subst ., ,$1))
MAJMINPATVER := $(shell cat version.txt)
MAJMINVER := $(call word-dot,$(MAJMINPATVER),1).$(call word-dot,$(MAJMINPATVER),2)
PATVER := $(call word-dot,$(MAJMINPATVER),3)


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

rebuild: clean build

build: fmt
	$(GO) build -o $(BUILD_BINARY) -v main.go


.PHONY: clean
clean:
	$(DELETE) -rf bin/ public/ zips/
	$(GO) clean -cache

.PHONY: fmt
fmt:
	$(GOFMT) -l -w $(SRC)

release:
	$(MAKE) release/linux GOARCH=amd64 & \
	$(MAKE) release/linux GOARCH=arm64 & \
	$(MAKE) release/linux GOARCH=arm & \
	$(MAKE) release/darwin & \
	wait

bindir = bin/$(subst release/,,$1)-$(GOARCH)
release/%: fmt
	@echo "build GOOS: $(subst release/,,$@) & GOARCH: $(GOARCH)"
	GOOS=$(subst release/,,$@) GOARCH=$(GOARCH) $(GO) build -ldflags="-s -w" -o $(call bindir,$@)/$(OS_BINARY) -v main.go
	mkdir -p zips
	zip -j zips/$(BINARY)-$(subst release/,,$@)-$(GOARCH).zip README.md $(call bindir,$@)/*

.PHONY: test
test: build
	$(GO) test -race ./...

.PHONY: test
coverage: build
	$(GO) test -covermode=atomic --race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out
