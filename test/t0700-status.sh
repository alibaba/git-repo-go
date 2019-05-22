#!/bin/sh

test_description="test 'git-repo status'"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -g all -u $manifest_url &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "project1 has untracked module1" '
	(
		cd work &&
		git-repo status
	) >actual &&
	cat >expect<<-EOF &&
	project projects/app1/                          (*** NO BRANCH ***)
	 --	module1/

	EOF
	test_cmp expect actual
'

test_expect_success "start new branch" '
	(
		cd work &&
		git-repo start --all jx/topic &&
		git-repo status
	) >actual &&
	cat >expect<<-EOF &&
	project projects/app1/                          branch jx/topic
	 --	module1/

	EOF
	test_cmp expect actual
'

test_expect_success "ignore module1" '
	(
		cd work/projects/app1 &&
		echo "/module1/" >.gitignore &&
		git add .gitignore &&
		git-repo status
	) >actual &&
	cat >expect<<-EOF &&
	project projects/app1/                          branch jx/topic
	 A-	.gitignore

	EOF
	test_cmp expect actual
'

test_expect_success "update .gitignore" '
	(
		cd work/projects/app1 &&
		echo "*.o" >>.gitignore &&
		git-repo status
	) >actual &&
	cat >expect<<-EOF &&
	project projects/app1/                          branch jx/topic
	 Am	.gitignore

	EOF
	test_cmp expect actual
'

test_expect_success "working directory clean" '
	(
		cd work/projects/app1 &&
		git add -u &&
		git commit -m "ignore submodule"
	) &&
	(
		cd work &&
		git-repo status
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	NOTE: nothing to commit (working directory clean)
	EOF
	test_cmp expect actual
'

test_done
