#!/bin/sh

test_description="test git-repo download: one remote, no tracking"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git clone "$main_repo_url" main
	)
'

test_expect_success "remove remote and fail to download" '
	(
		cd work/main &&
		git remote remove origin &&
		git checkout -b jx/topic
	)

'

test_expect_success "no remote and fail to download" '
	(
		cd work/main &&
		test_must_fail git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			12345
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	WARNING: no remote defined for project main
	FATAL: not remote tracking defined, and do not know where to download
	EOF
	test_cmp expect actual
'

test_expect_success "add remote origin and set review-url" '
	(
		cd work/main &&
		git remote add origin "$main_repo_url" &&
		git config remote.origin.review https://example.com
	)
'

test_expect_success "download and checkout" '
	(
		cd work/main &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			12345
	)
'

test_expect_success "after download HEAD is detached" '
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2 &&
		git show-ref | cut -c 42- | grep merge-requests
	) >out 2>&1 &&
	sed -e "s/(no branch)/Detached HEAD/g" out >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_done
