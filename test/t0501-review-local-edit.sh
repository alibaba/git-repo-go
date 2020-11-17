#!/bin/sh

test_description="upload test"

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

test_expect_success "install git pr alias command" '
	git-repo version &&
	git config alias.pr >actual &&
	git config alias.peer-review >>actual &&
	cat >expect <<-EOF &&
	repo upload --single
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "creant topic branch and set url to http protocol" '
	(
		cd work/main &&
		git checkout -b my/topic-test origin/master &&
		git config remote.origin.url \
			https://example.com/jiangxin/main.git
	)
'

test_expect_success "new commit in main project" '
	(
		cd work/main &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file"
	)
'

test_expect_success "confirm if has local edit" '
	(
		cd work &&
		echo "hack again" >>main/topic1.txt
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         7754413c5e1738581ef840d364d68cdc3fd19523
		to https://example.com (y/N)? Yes
		Uncommitted changes in jiangxin/main (did you forget to amend?):
		Continue uploading? (y/N) Yes
		NOTE: will execute command: git push ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		
		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git pr \
				--assume-yes \
				--no-edit \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"
		) >out 2>&1 &&
		sed -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_done
