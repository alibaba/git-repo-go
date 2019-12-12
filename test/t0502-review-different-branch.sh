#!/bin/sh

test_description="git peer-review on different branch"

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
	test_must_fail git-repo &&
	git config alias.peer-review >actual &&
	git config alias.pr >>actual &&
	git config alias.review >>actual &&
	cat >expect <<-EOF &&
	repo upload --single
	repo upload --single
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "update remote URL using http protocol" '
	(
		cd work/main &&
		git config remote.origin.url https://example.com/jiangxin/main.git
	)
'

test_expect_success "upload error: not in a branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: upload failed: not in a branch
		
		Please run command "git checkout -b <branch>" to create a new branch.
		EOF
		(
			cd main &&
			# git v1.7.10: "git checkout -q" is not really quiet.
			git checkout -q -b jx/topic1 origin/master >/dev/null &&
			echo hack >topic1.txt &&
			git add topic1.txt &&
			git commit -q -m "add topic1.txt" &&
			git checkout -q master^0 &&
			test_must_fail git peer-review \
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

test_expect_success "upload: pr --br <branch> to upload specific branch" '
	(
		cd work &&
		(
			cd main &&
			git peer-review \
				--br jx/topic1 \
				--assume-yes \
				--no-edit \
				--dryrun \
				--draft \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master (draft):
		  branch jx/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic1:refs/drafts/master/jx/topic1
		NOTE: with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic1 on refs/heads/jx/topic1, reason: review from jx/topic1 to master on https://example.com
		
		----------------------------------------------------------------------
		EOF
		test_cmp expect actual
	)
'

test_done
