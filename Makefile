GO=go
GOARCH ?= amd64
GOFMT=gofmt
DELETE=rm
BINARY=aardy
BIN_DIR=bin
BUILD_BINARY=$(BIN_DIR)/$(BINARY)
# go source files, ignore vendor directory
SRC := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
# current git version short-hash
VER := $(shell git rev-parse --short HEAD)
WATCH := (.go$$)

word-dot = $(word $2,$(subst ., ,$1))
MAJMINPATVER := $(shell cat version.txt)
MAJMINVER := $(call word-dot,$(MAJMINPATVER),1).$(call word-dot,$(MAJMINPATVER),2)
PATVER := $(call word-dot,$(MAJMINPATVER),3)

ifeq ($(CI),true)
  S3_BASEURL := s3://aardlabs-apps/cli/$(BINARY)
else
  S3_BASEURL := s3://aardlabs-apps-nonprod/cli/$(BINARY)
endif

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

build: fmt verinfo
	$(GO) build -o $(BUILD_BINARY) -v main.go
	cp misc/**/*.sh $(BIN_DIR)/

watch: $(BIN_DIR)/.reflex-installed
	reflex -r "$(WATCH)" -s -- make build
.PHONY: watch

.PHONY: clean
clean:
	$(DELETE) -rf bin/ public/ zips/
	$(GO) clean -cache

.PHONY: fmt
fmt:
	$(GOFMT) -l -w $(SRC)

release:
	mkdir -p public
ifeq ($(AWS_SECRET_ACCESS_KEY),)
  ifeq ($(CI),true)
	$(error AWS credentials not provided: Unable to sync with S3 ***)
  endif
else
# Download three previous versions to allow for binary diff...
	@for i in `seq 0 2`; do \
	  u=$(S3_BASEURL)/$(MAJMINVER).$$(( $(PATVER) - $$i ))/; \
	  echo "Attempting to sync $$u:"; \
	  aws s3 sync $$u public/; \
	done; true # ignore fails
endif
	$(MAKE) release/linux GOARCH=amd64 & \
	$(MAKE) release/linux GOARCH=arm64 & \
	$(MAKE) release/linux GOARCH=arm & \
	$(MAKE) release/darwin & \
	wait
ifneq ($(AWS_SECRET_ACCESS_KEY),)
# Upload the diffs and new version...
	aws s3 sync public/ $(S3_BASEURL)/
endif

bindir = bin/$(subst release/,,$1)-$(GOARCH)
release/%: fmt verinfo $(BIN_DIR)/.selfupdate-installed
	@echo "build GOOS: $(subst release/,,$@) & GOARCH: $(GOARCH)"
	GOOS=$(subst release/,,$@) GOARCH=$(GOARCH) $(GO) build -ldflags="-s -w" -o $(call bindir,$@)/$(BINARY) -v main.go
	cp misc/**/*.sh $(call bindir,$@)
	GOOS=$(subst release/,,$@) GOARCH=$(GOARCH) go-selfupdate $(call bindir,$@)/$(BINARY) $(MAJMINPATVER)
	mkdir -p zips
	zip -j zips/$(BINARY)-$(subst release/,,$@)-$(GOARCH).zip README.md $(call bindir,$@)/*

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

$(BIN_DIR)/.reflex-installed:
	@mkdir -p $(BIN_DIR)
	@cd /tmp && \
	  go get -u github.com/cespare/reflex && \
	  go install github.com/cespare/reflex@latest
	@touch $@

$(BIN_DIR)/.selfupdate-installed:
	@mkdir -p $(BIN_DIR)
	@cd /tmp && \
	  go get -u github.com/sanbornm/go-selfupdate/...
	@touch $@

FORCE: