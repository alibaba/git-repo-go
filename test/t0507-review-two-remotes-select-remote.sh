#!/bin/sh

test_description="single repo with two remotes, no origin"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	# checkout main.git and make it detached
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git clone $main_repo_url main
	)
'

test_expect_success "install git review aliases command" '
	git-repo version &&
	git config alias.pr >>actual &&
	cat >expect <<-EOF &&
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "commit in new topic branch" '
	(
		cd work/main &&
		# git v1.7.10: "git checkout -q" is not really quiet.
		git checkout -q -b jx/topic1 >/dev/null &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		git commit -q -m "add topic1.txt"
	)
'

test_expect_success "add two remote, one is origin" '
	(
		cd work/main &&
		git remote remove origin &&
		git remote add 1st "$main_repo_url" &&
		git fetch 1st &&
		git remote set-url 1st "https://example.com/jiangxin/main.git" &&
		git remote add 2nd "$main_repo_url" &&
		git fetch 2nd &&
		git remote set-url 2nd "https://example.com/jiangxin/main.git"
	)
'

test_expect_success "upload error: no tracking branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: no remote for branch '"'"'jx/topic1'"'"' of project '"'"'main'"'"' to push
		EOF
		(
			cd main &&
			test_must_fail git pr \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"

		) >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "upload --remote 1st --dest <branch>" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch jx/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic1:refs/for/master/jx/topic1
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic1 on refs/heads/jx/topic1, reason: review from jx/topic1 to master on https://example.com
		
		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git pr  --remote 1st --dest master \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"

		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "detach head" '
	(
		cd work/main &&
		git checkout -q HEAD^0 >/dev/null
	)
'

test_expect_success "upload --remote 1st --br <branch> --dest <branch>" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch jx/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic1:refs/for/master/jx/topic1
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic1 on refs/heads/jx/topic1, reason: review from jx/topic1 to master on https://example.com
		
		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git pr --remote 2nd --br jx/topic1 --dest master \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"

		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_done
