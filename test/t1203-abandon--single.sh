#!/bin/sh

test_description="test 'git-repo abandon --single'"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	mkdir work &&
	(
		cd work &&
		git clone $main_repo_url main
	)
'

test_expect_success "create branches" '
	(
		cd work/main &&
		test -d .git &&
		git checkout -b jx/topic1 origin/master &&
		git checkout -b jx/topic2 origin/master &&
		test_tick &&
		git commit --allow-empty -m "topic2: hack1" &&
		git checkout -b jx/topic3 origin/master &&
		test_tick &&
		git commit --allow-empty -m "topic3: hack1" &&
		test_tick &&
		git commit --allow-empty -m "topic3: hack2" &&
		git checkout -b jx/topic4 origin/master &&
		git checkout -b jx/topic5 origin/master &&
		echo hack >>README.md &&
		git add -u
	)
'

test_expect_success "git-repo abandon --single -b <branch>" '
	(
		cd work/main &&
		test -d .git &&
		git-repo abandon --single -b jx/topic2
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project ./
	  jx/topic2 ( 1 commit, Thu Apr 7 15:13:13 -0700 2005)
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon --single -b <branch>, by force" '
	(
		cd work/main &&
		test -d .git &&
		git-repo abandon --single -b jx/topic2 --force
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Abandoned branches
	------------------------------------------------------------------------------
	jx/topic2                 | . (was e10ee37)
	
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon --single --all" '
	(
		cd work/main &&
		test -d .git &&
		git-repo abandon --single --all
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Pruned branches (already merged)
	------------------------------------------------------------------------------
	jx/topic1                 | . (was 152dee6)
	
	jx/topic4                 | . (was 152dee6)
	
	master                    | . (was 152dee6)
	
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project ./
	  jx/topic3 ( 2 commits, Thu Apr 7 15:15:13 -0700 2005)
	* jx/topic5
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon --single --all, by force" '
	(
		cd work/main &&
		test -d .git &&
		git-repo abandon --single --all --force
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Abandoned branches
	------------------------------------------------------------------------------
	jx/topic3                 | . (was 63072c1)
	
	jx/topic5                 | . (was 152dee6)
	
	EOF
	test_cmp expect actual
'

test_done
