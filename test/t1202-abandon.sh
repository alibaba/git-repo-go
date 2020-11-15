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
		(
			cd drivers/driver-1 &&
			test_tick &&
			git commit --allow-empty -m "topic1 on driver1: hack1"
		) &&
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

test_expect_success "git-repo abandon a single branch" '
	(
		cd work &&
		git-repo abandon -b jx/topic1
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Pruned branches (already merged)
	------------------------------------------------------------------------------
	jx/topic1                 | drivers/driver-2      (was df3d4c6)
	                          | main                  (was 4d13a6c)
	                          | projects/app1         (was 2fdfd9b)
	                          | projects/app1/module1 (was 8fc882d)
	                          | projects/app2         (was 98dc74a)
	
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project drivers/driver-1/
	  jx/topic1 ( 1 commit, Thu Apr 7 15:14:13 -0700 2005)
		EOF
		test_cmp expect actual
'

test_expect_success "git-repo abandon a single branch, twice" '
	(
		cd work &&
		git-repo abandon -b jx/topic1
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	ERROR: project main> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app1> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app1/module1> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app2> fail to resolve refs/heads/jx/topic1
	ERROR: project drivers/driver-2> fail to resolve refs/heads/jx/topic1
	Pending branches (which have unmerged commits, leave it as is)
	------------------------------------------------------------------------------
	Project drivers/driver-1/
	  jx/topic1 ( 1 commit, Thu Apr 7 15:14:13 -0700 2005)
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon a single branch, force" '
	(
		cd work &&
		git-repo abandon -b jx/topic1 --force
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	ERROR: project main> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app1> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app1/module1> fail to resolve refs/heads/jx/topic1
	ERROR: project projects/app2> fail to resolve refs/heads/jx/topic1
	ERROR: project drivers/driver-2> fail to resolve refs/heads/jx/topic1
	Abandoned branches
	------------------------------------------------------------------------------
	jx/topic1                 | drivers/driver-1      (was ce87c62)
	
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon selected projects, force" '
	(
		cd work &&
		git-repo abandon --all --force main projects/app2
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Abandoned branches
	------------------------------------------------------------------------------
	jx/topic2                 | main          (was 4d13a6c)
	                          | projects/app2 (was 98dc74a)
	
	jx/topic3                 | main          (was 4d13a6c)
	                          | projects/app2 (was 98dc74a)
	
	jx/topic4                 | main          (was 4d13a6c)
	                          | projects/app2 (was 98dc74a)

	EOF
	test_cmp expect actual
'

test_expect_success "git-repo abandon all projects" '
	(
		cd work &&
		git-repo abandon --all --force
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Abandoned branches
	------------------------------------------------------------------------------
	jx/topic2                 | drivers/driver-1      (was e05767b)
	                          | drivers/driver-2      (was 8ff4cd2)
	                          | projects/app1         (was 2fdfd9b)
	                          | projects/app1/module1 (was f556e17)
	
	jx/topic3                 | drivers/driver-1      (was faa6f5c)
	                          | drivers/driver-2      (was df3d4c6)
	                          | projects/app1         (was 2fdfd9b)
	                          | projects/app1/module1 (was 8fc882d)
	
	jx/topic4                 | drivers/driver-1      (was faa6f5c)
	                          | drivers/driver-2      (was df3d4c6)
	                          | projects/app1         (was 2fdfd9b)
	                          | projects/app1/module1 (was 3eabae6)

	EOF
	test_cmp expect actual
'

test_done
