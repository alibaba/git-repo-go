#!/bin/sh

test_description="start used with --mirror"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init" '
	(
		cd work &&
		git-repo init --mirror -u $manifest_url -g all -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "fail: cannot run start used with --mirror" '
	(
		cd work &&
		test_must_fail git-repo start --all my/topic1
	) >actual 2>&1 &&
	cat >expect <<-EOF &&
	FATAL: cannot run in a mirror
	EOF
	test_cmp expect actual
'

test_done
