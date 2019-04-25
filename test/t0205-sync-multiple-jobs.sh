#!/bin/sh

test_description="git-repo sync multiple jobs test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "init with multiple jobs" '
	(
		cd work &&
		git-repo init -u $manifest_url
	)
'

test_expect_success "git-repo sync (-n), default jobs" '
	(
		cd work &&
		git-repo sync -n
	)
'

test_expect_success "git-repo sync (-n), 10 jobs" '
	(
		cd work &&
		git-repo sync -n -j 10
	)
'

test_expect_success "git-repo sync (-n), 1 job" '
	(
		cd work &&
		git-repo sync -n -j 1
	)
'

test_expect_success "git-repo sync (-n), 0 job" '
	(
		cd work &&
		git-repo sync -n -j 0
	)
'


test_done
