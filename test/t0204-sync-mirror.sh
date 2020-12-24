#!/bin/sh

test_description="sync overwrites dirty workdir if workdir is in detached mode"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync in mirror mode" '
	(
		cd work &&
		git-repo init -u $manifest_url -g all --mirror -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "no project repositories in .repo" '
	(
		cd work &&
		test -d .repo/manifests.git &&
		test ! -d .repo/projects &&
		test ! -d .repo/project-objects
	)
'

test_expect_success "project repositories in workspace" '
	(
		cd work &&
		test -d hello/manifests.git &&
		test -d drivers/driver1.git &&
		test -d drivers/driver2.git &&
		test -d main.git &&
		test -d project1.git &&
		test -d project2.git &&
		test -d project1/module1.git
	)
'

test_expect_success "git-repo init with tag" '
	(
		cd work &&
		git-repo init -b refs/tags/v0.2 &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_done
