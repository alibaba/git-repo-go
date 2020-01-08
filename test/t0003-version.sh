#!/bin/sh

test_description="test git-repo version"

. ./lib/sharness.sh

test_expect_success "git-repo version output test" '
	git-repo version >out &&
	grep "^git-repo version" out >expect &&
	test -s expect
'

test_expect_success "check version number format" '
	grep "^git-repo version [0-9]\+\.[0-9]\+\.[0-9]\+" out >actual &&
	test_cmp expect actual
'

test_done
