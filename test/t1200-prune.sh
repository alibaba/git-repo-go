#!/bin/sh

test_description="test 'git-repo prune'"

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

test_expect_success "create branches" '
	(
		cd work &&
		git-repo start --all jx/topic1 &&
		git-repo start --all jx/topic2 &&
		(
			cd drivers/driver-1 &&
			test_tick &&
			git commit --allow-empty -m "topic2 on driver1: hack1"
		) &&
		(
			cd drivers/driver-2 &&
			test_tick &&
			git commit --allow-empty -m "topic2 on driver2: hack1" &&
			test_tick &&
			git commit --allow-empty -m "topic2 on driver2: hack2"
		) &&
		(
			cd projects/app1/module1 &&
			test_tick &&
			git commit --allow-empty -m "topic2 on mod1: hack1" &&
			test_tick &&
			git commit --allow-empty -m "topic2 on mod1: hack2" &&
			test_tick &&
			git commit --allow-empty -m "topic2 on mod1: hack3"
		)
	)
'

test_expect_success "create branches and dirty worktree" '
	(
		cd work &&
		git-repo start --all jx/topic3 &&
		git-repo start --all jx/topic4 &&
		(
			cd drivers/driver-1 &&
			echo hack >>README.md &&
			git add -u
		) &&
		(
			cd drivers/driver-2 &&
			git checkout HEAD^0 &&
			test_tick &&
			git commit --allow-empty -m "topic4 on driver2: hack1" &&
			echo hack >>README.md
		) &&
		(
			cd projects/app1/module1 &&
			test_tick &&
			git commit --allow-empty -m "topic4 on mod1: hack1" &&
			test_tick &&
			git commit --allow-empty -m "topic4 on mod1: hack2" &&
			test_tick &&
			git commit --allow-empty -m "topic4 on mod1: hack3"
		)
	)
'

test_expect_success "git-repo prune all" '
	(
		cd work &&
		git-repo prune
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	ERROR: cannot find tracking branch refs/heads/jx/topic4 of projects/app1/module1: revision refs/remotes/aone/jx/topic4 in project1/module1 not found
	Pruned branches (already merged)
	------------------------------------------------------------------------------
	jx/topic1                 | drivers/driver-1      (was 69d4c01)
	                          | drivers/driver-2      (was 4f58351)
	                          | main                  (was 152dee6)
	                          | projects/app1         (was eac322d)
	                          | projects/app1/module1 (was 2be33cb)
	                          | projects/app2         (was 927fd5d)
	
	jx/topic2                 | main                  (was 152dee6)
	                          | projects/app1         (was eac322d)
	                          | projects/app2         (was 927fd5d)
	
	jx/topic3                 | drivers/driver-1      (was 69d4c01)
	                          | drivers/driver-2      (was 4f58351)
	                          | main                  (was 152dee6)
	                          | projects/app1         (was eac322d)
	                          | projects/app1/module1 (was 2be33cb)
	                          | projects/app2         (was 927fd5d)
	
	jx/topic4                 | drivers/driver-2      (was 4f58351)
	                          | main                  (was 152dee6)
	                          | projects/app1         (was eac322d)
	                          | projects/app2         (was 927fd5d)
	
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project drivers/driver-1/
	  jx/topic2 ( 1 commit, Thu Apr 7 15:13:13 -0700 2005)
	* jx/topic4
	
	Project drivers/driver-2/
	  jx/topic2 ( 2 commits, Thu Apr 7 15:14:13 -0700 2005)
	
	Project projects/app1/module1/
	  jx/topic2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo prune some projects" '
	(
		cd work &&
		git-repo prune main drivers/driver-1 projects/app1
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project drivers/driver-1/
	  jx/topic2 ( 1 commit, Thu Apr 7 15:13:13 -0700 2005)
	* jx/topic4
	EOF
	test_cmp expect actual
'
	
test_done
