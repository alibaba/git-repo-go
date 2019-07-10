#!/bin/sh

test_description="test 'git-repo download' basic"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -u $manifest_url &&
		git-repo sync \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		git-repo start --all jx/topic
	)
'

test_expect_success "download and checkout" '
	(
		cd work &&
		git-repo download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			main 12345
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2 &&
		git show-ref | cut -c 42- | grep merge-requests
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_expect_success "download again with already merged notice" '
	(
		cd work &&
		git-repo download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			main 12345
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	NOTE: [main] change 12345 has already been merged
	EOF
	test_cmp expect actual &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2
		git show-ref | cut -c 42- | grep merge-requests
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_expect_success "restore using sync and start again" '
	(
		cd work &&
		git-repo sync --detach \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		git-repo start --all jx/topic
	)
'

test_expect_success "download using cherry-pick" '
	(
		cd work &&
		git-repo download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			--cherry-pick main 12345
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2
		git show-ref | cut -c 42- | grep merge-requests
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: jx/topic
	    New topic
	    Version 2.0.0-dev
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_expect_success "restore using sync and start again" '
	(
		cd work &&
		git-repo sync --detach \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		git-repo start --all jx/topic
	)
'

test_expect_success "download failed using ff-only" '
	(
		cd work &&
		test_must_fail git-repo download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			--ff-only main 12345
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	fatal: Not possible to fast-forward, aborting.
	Error: exit status 128
	EOF
	test_cmp expect actual
'

test_done
