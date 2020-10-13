TARGETS := git-repo
PKG := github.com/alibaba/git-repo-go
VENDOR_EXISTS=$(shell test -d vendor && echo 1 || echo 0)
ifeq ($(VENDOR_EXISTS), 1)
    GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor
    GOTEST := GO111MODULE=on go test -mod=vendor
else
    GOBUILD := GO111MODULE=on CGO_ENABLED=0 go build
    GOTEST := GO111MODULE=on go test
endif

ifeq ($(shell uname), Darwin)
    TAR=gtar
else
    TAR=tar
endif

GOBUILD_LINUX_64 := env GOOS=linux GOARCH=amd64 $(GOBUILD)
GOBUILD_LINUX_32 := env GOOS=linux GOARCH=386 $(GOBUILD)
GOBUILD_WINDOWS_64 := env GOOS=windows GOARCH=amd64 $(GOBUILD)
GOBUILD_WINDOWS_32 := env GOOS=windows GOARCH=386 $(GOBUILD)
GOBUILD_MAC_64 := env GOOS=darwin GOARCH=amd64 $(GOBUILD)
GOBUILD_MAC_32 := env GOOS=darwin GOARCH=386 $(GOBUILD)

BUILD_RELEASE_FLAG=-ldflags "-s -w"

SHA256SUM=shasum -a 256
GPGSIGN=gpg -sba -u Alibaba
# Returns a list of all non-vendored (local packages)
LOCAL_PACKAGES = $(shell go list ./... | grep -v -e '^$(PKG)/vendor/')
LOCAL_GO_FILES = $(shell find -L $BUILD_DIR  -name "*.go" -not -path "$(PKG_BUILD_DIR)/vendor/*" -not -path "$(PKG_BUILD_DIR)/_build/*")
#UPX = upx
UPX = echo Disabled upx

define message
	@echo "### $(1)"
endef

all: $(TARGETS)

REPO-VERSION-FILE: FORCE
	$(call message,Generate version file)
	@/bin/sh ./REPO-VERSION-GEN
-include REPO-VERSION-FILE

# Define LDFLAGS after include of REPO-VERSION-FILE
LDFLAGS := -ldflags "-X $(PKG)/version.Version=$(REPO_VERSION)"
RELEASE_LDFLAGS := -ldflags "-X $(PKG)/version.Version=$(REPO_VERSION) -s -w"

git-repo: $(shell find . -name '*.go') | REPO-VERSION-FILE
	$(call message,Building $@)
	$(GOBUILD) $(LDFLAGS) -o $@

golint:
	$(call message,Testing git-repo using golint for coding style)
	@golint $(LOCAL_PACKAGES)

test: golint ut it

ut: $(TARGETS)
	$(call message,Testing git-repo for unit tests)
	$(GOTEST) $(PKG)/...

it: $(TARGETS)
	$(call message,Testing git-repo for integration tests)
	@make -C test

it-no-prove: $(TARGETS)
	$(call message,Testing git-repo for integration tests (not using prove))
	@make -C test test

version-yml: REPO-VERSION-FILE
	@mkdir -p _build
	@echo "production: $(REPO_VERSION)" > _build/version.yml
	@echo "test: $(REPO_VERSION)" >> _build/version.yml

release: macOS Linux Windows

Linux: Linux-64 Linux-32
Linux-64: _build/$(REPO_VERSION)/Linux-64/git-repo
_build/$(REPO_VERSION)/Linux-64/git-repo: FORCE
	@$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_LINUX_64) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		$(TAR) -zcvf ../git-repo-$(REPO_VERSION)-Linux-64.tar.gz --transform "s/^\./git-repo-$(REPO_VERSION)-Linux-64/" .)

Linux-32: _build/$(REPO_VERSION)/Linux-32/git-repo
_build/$(REPO_VERSION)/Linux-32/git-repo: FORCE
	$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_LINUX_32) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		$(TAR) -zcvf ../git-repo-$(REPO_VERSION)-Linux-32.tar.gz --transform "s/^\./git-repo-$(REPO_VERSION)-Linux-32/" .)

Windows: Windows-64 Windows-32
Windows-64: _build/$(REPO_VERSION)/git-repo-$(REPO_VERSION)-Windows-64/git-repo.exe
_build/$(REPO_VERSION)/git-repo-$(REPO_VERSION)-Windows-64/git-repo.exe: FORCE
	$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_WINDOWS_64) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		cd .. && \
		zip -r git-repo-$(REPO_VERSION)-Windows-64.zip git-repo-$(REPO_VERSION)-Windows-64/)

Windows-32: _build/$(REPO_VERSION)/git-repo-$(REPO_VERSION)-Windows-32/git-repo.exe
_build/$(REPO_VERSION)/git-repo-$(REPO_VERSION)-Windows-32/git-repo.exe: FORCE
	$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_WINDOWS_32) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		cd .. && \
		zip -r git-repo-$(REPO_VERSION)-Windows-32.zip git-repo-$(REPO_VERSION)-Windows-32/)

macOS: macOS-64
macOS-64: _build/$(REPO_VERSION)/macOS-64/git-repo
_build/$(REPO_VERSION)/macOS-64/git-repo: FORCE
	$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_MAC_64) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		$(TAR) -zcvf ../git-repo-$(REPO_VERSION)-macOS-64.tar.gz --transform "s/^\./git-repo-$(REPO_VERSION)-macOS-64/" .)

# go 1.15 no longer support build of macOS-32
macOS-32: _build/$(REPO_VERSION)/macOS-32/git-repo
_build/$(REPO_VERSION)/macOS-32/git-repo: FORCE
	$(call message,Building $@)
	@mkdir -p $(shell dirname $@)
	$(GOBUILD_MAC_32) $(RELEASE_LDFLAGS) -o $@
	$(UPX) $@
	(cd $(shell dirname $@) && \
		$(SHA256SUM) $(shell basename $@) >$(shell basename $@).sha256 && \
		$(GPGSIGN) -o $(shell basename $@).sha256.gpg $(shell basename $@).sha256 && \
		$(TAR) -zcvf ../git-repo-$(REPO_VERSION)-macOS-32.tar.gz --transform "s/^\./git-repo-$(REPO_VERSION)-macOS-32/" .)

index:
	$(call message,Building $@)
	@mkdir -p _build/$(REPO_VERSION)/
	pandoc -s -f markdown -t html --metadata title="CHANGELOG of git-repo" -o _build/$(REPO_VERSION)/index.html CHANGELOG.md

clean:
	$(call message,Cleaning $(TARGETS))
	@rm -f $(TARGETS)
	@rm -f REPO-VERSION-FILE

.PHONY: test clean
.PHONY: FORCE
.PHONY: version-yml index
.PHONY: release
.PHONY: ut it it-no-prove
