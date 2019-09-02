#!/bin/sh

test_description="git-repo init --single test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "cannot run init --single" '
	(
		cd work &&
		test_must_fail git-repo init --single -u $manifest_url
	) >actual 2>&1 &&
	cat >expect <<-EOF &&
	FATAL: cannot run in single mode
	EOF
	test_cmp expect actual
'

test_done
