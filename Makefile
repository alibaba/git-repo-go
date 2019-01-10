TARGETS := git-repo
#GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-extldflags '-static'"
# GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-X main.Version=$(VERSION) -extldflags '-static'"
GOBUILD := go build
VERSION := $(shell git describe --always --tag --dirty)-$(shell date -u +%Y%m%d.%H%M%S)

# Returns a list of all non-vendored (local packages)
LOCAL_PACKAGES = $(shell go list ./... | grep -v -e '^$(PKG)/vendor/')
LOCAL_GO_FILES = $(shell find -L $BUILD_DIR  -name "*.go" -not -path "$(PKG_BUILD_DIR)/vendor/*" -not -path "$(PKG_BUILD_DIR)/_build/*")

define message
	@echo "### $(1)"
endef

all: $(TARGETS)

git-repo: $(shell find . -name '*.go')
	$(call message,Building $@)
	$(GOBUILD)

test: $(TARGETS)
	$(call message,Testing git-repo using golint for coding style)
	@golint ./...
	$(call message,Testing git-repo for unit tests)
	@go test $(LOCAL_PACKAGES)
	$(call message,Testing git-repo for integration tests)
	@make -C test

clean:
	$(call message,Cleaning $(TARGETS))
	@rm -f $(TARGETS)

.PHONY: test clean
