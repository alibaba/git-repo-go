TARGETS := git-repo
PKG := code.alibaba-inc.com/force/git-repo
#GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-extldflags '-static'"
# GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -tags netgo -ldflags "-X main.Version=$(VERSION) -extldflags '-static'"
GOBUILD := CGO_ENABLED=0 go build
GOBUILD_LINUX_64 := env GOOS=linux GOARCH=amd64 $(GOBUILD)
GOBUILD_LINUX_32 := env GOOS=linux GOARCH=386 $(GOBUILD)
GOBUILD_WINDOWS_64 := env GOOS=windows GOARCH=amd64 $(GOBUILD)
GOBUILD_WINDOWS_32 := env GOOS=windows GOARCH=386 $(GOBUILD)
GOBUILD_MAC_64 := env GOOS=darwin GOARCH=amd64 $(GOBUILD)
GOBUILD_MAC_32 := env GOOS=darwin GOARCH=386 $(GOBUILD)

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
	$(GOBUILD)  -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

golint:
	$(call message,Testing git-repo using golint for coding style)
	@golint ./...

test: golint $(TARGETS)
	$(call message,Testing git-repo for unit tests)
	@go test $(LOCAL_PACKAGES)
	$(call message,Testing git-repo for integration tests)
	@make -C test

linux-64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/linux-amd64
	$(GOBUILD_LINUX_64) -o target/linux-amd64/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

linux-32: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/linux-i386
	$(GOBUILD_LINUX_32) -o target/linux-i386/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

windows-64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/windows-x86_64
	$(GOBUILD_WINDOWS_64) -o target/windows-x86_64/git-repo.exe -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

windows-32: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/windows-i386
	$(GOBUILD_WINDOWS_32) -o target/windows-i386/git-repo.exe -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

mac-64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/Mac-x86_64
	$(GOBUILD_MAC_64) -o target/Mac-x86_64/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

mac-32: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p target/Mac-i386
	$(GOBUILD_MAC_32) -o target/Mac-i386/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"

clean:
	$(call message,Cleaning $(TARGETS))
	@rm -f $(TARGETS)
	@rm -f REPO-VERSION-FILE

.PHONY: test clean
.PHONY: FORCE
