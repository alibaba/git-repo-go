#!/bin/sh

test_description="test git-repo version"

. ./lib/sharness.sh

test_expect_success "git-repo version format checking" '
	git-repo version 2>/dev/null | grep "^git-repo version [0-9]\+\.[0-9]\+\.[0-9]\+"
'

test_done
