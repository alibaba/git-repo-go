#!/bin/sh

test_description="sync will update and reload manifests project"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync with history version of manifests" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.1 &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		git -C .repo/manifests config branch.default.merge refs/heads/master &&
		git -C .repo/manifests config branch.default.remote origin
	)
'

test_expect_success "manifests version: 0.1.0" '
	(
		cd work &&
		cat >expect<<-EOF &&
		manifests: Version 0.1
		EOF
		git -C .repo/manifests log -1 --pretty="manifests: %s" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "VERSION: 0.1.0" '
	(
		cd work &&
		test -d main &&
		test -f VERSION &&
		test -L Makefile &&
		cat >expect<<-EOF &&
		v0.1.0
		EOF
		test_cmp expect VERSION
	)
'

test_expect_success "project.list: 1 project" '
	(
		cd work &&
		cat >expect<<-EOF &&
		main
		EOF
		test_cmp expect .repo/project.list
	)
'

test_expect_success "verify checkout commits of v0.1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		main: Version 0.1.0
		EOF
		git -C main log -1 --pretty="main: %s" >actual &&
		test_cmp expect actual
	)
'


test_expect_success "git-repo sync again, upgrade manifest first" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests version: 2.0" '
	(
		cd work &&
		cat >expect<<-EOF &&
		manifests: Version 2.0
		EOF
		git -C .repo/manifests log -1 --pretty="manifests: %s" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "project.list: 5 project" '
	(
		cd work &&
		cat >expect<<-EOF &&
		drivers/driver-1
		main
		projects/app1
		projects/app1/module1
		projects/app2
		EOF
		test_cmp expect .repo/project.list
	)
'

test_expect_success "verify checkout commits of master" '
	(
		cd work &&
		cat >expect<<-EOF &&
		drivers/dirver1: Version 2.0.0-dev
		main: Version 2.0.0-dev
		projects/app1: Version 2.0.0-dev
		projects/app1/module1: Version 1.0.0
		projects/app2: Version 2.0.0-dev
		EOF
		git -C drivers/driver-1 log -1 --pretty="drivers/dirver1: %s" >actual &&
		git -C main log -1 --pretty="main: %s" >>actual &&
		git -C projects/app1 log -1 --pretty="projects/app1: %s" >>actual &&
		git -C projects/app1/module1 log -1 --pretty="projects/app1/module1: %s" >>actual &&
		git -C projects/app2 log -1 --pretty="projects/app2: %s" >>actual &&
		test_cmp expect actual
	)
'

test_done
