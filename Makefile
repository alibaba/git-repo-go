TARGETS := git-repo
PKG := code.alibaba-inc.com/force/git-repo
#GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-extldflags '-static'"
# GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-X main.Version=$(VERSION) -extldflags '-static'"
GOBUILD := go build

# Returns a list of all non-vendored (local packages)
LOCAL_PACKAGES = $(shell go list ./... | grep -v -e '^$(PKG)/vendor/')
LOCAL_GO_FILES = $(shell find -L $BUILD_DIR  -name "*.go" -not -path "$(PKG_BUILD_DIR)/vendor/*" -not -path "$(PKG_BUILD_DIR)/_build/*")

define message
	@echo "### $(1)"
endef

all: $(TARGETS)

REPO-VERSION-FILE: FORCE
	@/bin/sh ./REPO-VERSION-GEN
-include REPO-VERSION-FILE

git-repo: $(shell find . -name '*.go')
	$(call message,Building $@)
	$(GOBUILD)  -ldflags "-X $(PKG)/version.Version=$(REPO_VERSION)"

golint:
	$(call message,Testing git-repo using golint for coding style)
	@golint ./...

test: golint $(TARGETS)
	$(call message,Testing git-repo for unit tests)
	@go test $(LOCAL_PACKAGES)
	$(call message,Testing git-repo for integration tests)
	@make -C test

clean:
	$(call message,Cleaning $(TARGETS))
	@rm -f $(TARGETS)
	@rm -f REPO-VERSION-FILE

.PHONY: test clean
.PHONY: FORCE
