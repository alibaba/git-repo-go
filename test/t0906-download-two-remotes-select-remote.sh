#!/bin/sh

test_description="test git-repo download: two remotes, select origin using --remote"

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
		git remote rename origin 1st &&
		git config remote.1st.review https://example.com &&
		git remote add 2nd "$main_repo_url" &&
		git config remote.2nd.review https://example.com &&
		git checkout -b jx/topic
	)
'

test_expect_success "fail to download without --remote" '
	(
		cd work/main &&
		git-repo download --single \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			12345
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	ERROR: no tracking remote defined, and don'"'"'t know where to download from.
	ERROR: please try to use --remote option for download
	EOF
	test_cmp expect actual
'

test_expect_success "download with --remote" '
	(
		cd work/main &&
		git-repo download --single \
			--remote 2nd \
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
	sed -e "s/(no branch)/Detached HEAD/g" out >actual &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_done
