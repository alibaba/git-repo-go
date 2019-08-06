#!/bin/sh

test_description="test git-repo download: two remotes, default origin'"

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

test_expect_success "add two remotes" '
	(
		cd work/main &&
		git config remote.origin.review https://example.com &&
		git remote add another "$main_repo_url" &&
		git config remote.another.review https://example.com &&
		git checkout -b jx/topic
	)

'

test_expect_success "default download from origin" '
	(
		cd work/main &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			12345
	) 2>&1 | head -1 >actual &&
	cat >expect<<-EOF &&
	WARNING: no tracking remote defined, try to download from origin
	EOF
	test_cmp expect actual &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2 &&
		git show-ref | cut -c 42- | grep merge-requests
	) | sed -e "s/(no branch)/Detached HEAD/g" >actual 2>&1 &&
	cat >expect<<-EOF &&
	Branch: jx/topic
	    Version 2.0.0-dev
	    Version 1.0.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_done
