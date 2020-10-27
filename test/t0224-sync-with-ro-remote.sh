#!/bin/sh

test_description="sync with read-only remote"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync using remote-ro.xml" '
	(
		cd work &&
		git-repo init -u $manifest_url -g all -m remote-ro.xml &&
		git-repo sync --no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

# TODO: add other test cases for remote with no review attribute.

test_done
