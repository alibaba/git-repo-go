#!/bin/sh

test_description="start new branch test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init" '
	(
		cd work &&
		git-repo init -u $manifest_url -g all -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "create branch: my/topic1" '
	(
		cd work &&
		git-repo start --all my/topic1
	)
'

test_expect_success "check current branch" '
	(
		cd work &&
		echo "main: my/topic1" >expect &&
		(printf "main: " && cd main && git_current_branch) >actual &&
		test_cmp expect actual &&
		echo "driver1: my/topic1" >expect &&
		(printf "driver1: " && cd drivers/driver-1 && git_current_branch) >actual &&
		test_cmp expect actual &&
		echo "driver2: my/topic1" >expect &&
		(printf "driver2: " && cd drivers/driver-2 && git_current_branch) >actual &&
		test_cmp expect actual &&
		echo "app1: my/topic1" >expect &&
		(printf "app1: " && cd projects/app1 && git_current_branch) >actual &&
		test_cmp expect actual &&
		echo "app2: my/topic1" >expect &&
		(printf "app2: " && cd projects/app2 && git_current_branch) >actual &&
		test_cmp expect actual &&
		echo "module1: my/topic1" >expect &&
		(printf "module1: " && cd projects/app1/module1 && git_current_branch) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "check tracking branch" '
	(
		cd work &&
		show_all_repo_branch_tracking >actual &&
		cat >expect <<-EOF &&
		## main
		   my/topic1 => refs/heads/Maint
		## projects/app1
		   my/topic1 => refs/heads/Maint
		## projects/app1/module1
		   my/topic1 => 
		## projects/app2
		   my/topic1 => refs/heads/Maint
		## drivers/driver-1
		   my/topic1 => refs/heads/Maint
		## drivers/driver-2
		   my/topic1 => refs/heads/Maint
		EOF
		test_cmp expect actual
	)
'

test_expect_success "create new commit in app1" '
	(
		cd work/projects/app1 &&
		git commit --allow-empty -m "my/topic1: test-commit" 
	)
'

test_expect_success "create new branch my/topic2" '
	(
		cd work &&
		git-repo start --all my/topic2
	)
'

test_expect_success "check current commit of app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		my/topic2> Version 1.0-dev
		EOF
		(
			printf "$(cd projects/app1 && git_current_branch)> " &&
			cd projects/app1 &&
			git log -1 --pretty="%s"
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "switch back to my/topic1" '
	(
		cd work &&
		git-repo start --all my/topic1
	)
'

test_expect_success "check current commit of app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		my/topic1> my/topic1: test-commit
		EOF
		(
			printf "$(cd projects/app1 && git_current_branch)> " &&
			cd projects/app1 &&
			git log -1 --pretty="%s"
		) >actual &&
		test_cmp expect actual
	)
'

test_done
