#!/bin/sh

test_description="Example test cases for git-repo"

. ./lib/sharness.sh

test_expect_success "Execute git-repo" '
	git-repo
'

test_done
