#!/bin/sh

test_description="test 'git-repo download' basic"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git clone $main_repo_url main &&
		cd main &&
		git checkout -b jx/topic origin/master
	)
'

test_expect_success "bad review url" '
	(
		cd work/main &&
		test_must_fail git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			12345/1
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	WARNING: no remote defined for project main
	FATAL: not remote tracking defined, and do not know where to download
	EOF
	test_cmp expect actual
'

test_expect_success "set review-url" '
	(
		cd work/main &&
		git config remote.origin.review https://example.com
	)
'

test_expect_success "download and checkout" '
	(
		cd work/main &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			12345/1
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2 &&
		git show-ref | cut -c 42- | grep changes
	) | sed -e "s/(no branch)/Detached HEAD/g" >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/changes/45/12345/1
	EOF
	test_cmp expect actual
'

test_expect_success "download again with already merged notice" '
	(
		cd work/main &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			. 12345/1
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	NOTE: [main] change 12345/1 has already been merged
	EOF
	test_cmp expect actual &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2
		git show-ref | cut -c 42- | grep changes
	) | sed -e "s/(no branch)/Detached HEAD/g" >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/changes/45/12345/1
	EOF
	test_cmp expect actual
'

test_expect_success "download using cherry-pick" '
	(
		cd work/main &&
		git checkout jx/topic &&
		git reset --quiet --hard origin/master &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			--cherry-pick \
			12345/2
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2
		git show-ref | cut -c 42- | grep changes
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: jx/topic
	    New topic
	    Version 2.0.0-dev
	refs/changes/45/12345/1
	refs/changes/45/12345/2
	EOF
	test_cmp expect actual
'

test_expect_success "download failed using ff-only" '
	(
		cd work/main &&
		git reset --quiet --hard origin/master &&
		test_must_fail git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			--ff-only \
			12345/2
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	fatal: Not possible to fast-forward, aborting.
	Error: exit status 128
	EOF
	test_cmp expect actual
'

test_expect_success "alias download command (cherry-pick)" '
	(
		cd work/main &&
		git checkout jx/topic &&
		git reset --quiet --hard origin/master &&
		git download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			--cherry-pick \
			12345/2
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2
		git show-ref | cut -c 42- | grep changes
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: jx/topic
	    New topic
	    Version 2.0.0-dev
	refs/changes/45/12345/1
	refs/changes/45/12345/2
	EOF
	test_cmp expect actual
'

test_done
