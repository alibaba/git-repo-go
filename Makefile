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

version-yml: REPO-VERSION-FILE
	@mkdir -p _build
	@echo "production: $(REPO_VERSION)" > _build/version.yml
	@echo "test: $(REPO_VERSION)" >> _build/version.yml

linux-amd64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/linux/amd64
	$(GOBUILD_LINUX_64) -o _build/linux/amd64/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

linux-386: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/linux/386
	$(GOBUILD_LINUX_32) -o _build/linux/386/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

windows-amd64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/windows/amd64
	$(GOBUILD_WINDOWS_64) -o _build/windows/amd64/git-repo.exe -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

windows-386: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/windows/386
	$(GOBUILD_WINDOWS_32) -o _build/windows/386/git-repo.exe -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

darwin-amd64: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/darwin/amd64
	$(GOBUILD_MAC_64) -o _build/darwin/amd64/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

darwin-386: $(shell find . -name '*.go')
	$(call message,Building $@)
	@mkdir -p _build/darwin/386
	$(GOBUILD_MAC_32) -o _build/darwin/386/git-repo -ldflags "-X $(PKG)/versions.Version=$(REPO_VERSION)"
	@make version-yml

clean:
	$(call message,Cleaning $(TARGETS))
	@rm -f $(TARGETS)
	@rm -f REPO-VERSION-FILE

.PHONY: test clean
.PHONY: FORCE
.PHONY: version-yml
