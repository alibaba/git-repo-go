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

test_expect_success "install git review alias command" '
	git-repo --version &&
	git config alias.review >actual &&
	cat >expect <<-EOF &&
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "creant topic branch and set url to http protocol" '
	(
		cd work &&
		git -C main checkout -b my/topic-test origin/master &&
		git -C main config remote.origin.url \
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
		         725d95a019384bec328ae99b1565f0eec25d02e5
		to https://example.com (y/N)? Yes
		Uncommitted changes in jiangxin/main (did you forget to amend?):
		Continue uploading? (y/N) Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
		test_cmp expect actual
	)
'

test_done
