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
		git-repo init -u $manifest_url -g all -b maint &&
		git-repo sync
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
		echo "main: refs/heads/my/topic1" >expect &&
		(printf "main: " && git -C main symbolic-ref HEAD) >actual &&
		test_cmp expect actual &&
		echo "driver1: refs/heads/my/topic1" >expect &&
		(printf "driver1: " && git -C drivers/driver-1 symbolic-ref HEAD) >actual &&
		test_cmp expect actual &&
		echo "driver2: refs/heads/my/topic1" >expect &&
		(printf "driver2: " && git -C drivers/driver-2 symbolic-ref HEAD) >actual &&
		test_cmp expect actual &&
		echo "app1: refs/heads/my/topic1" >expect &&
		(printf "app1: " && git -C projects/app1 symbolic-ref HEAD) >actual &&
		test_cmp expect actual &&
		echo "app2: refs/heads/my/topic1" >expect &&
		(printf "app2: " && git -C projects/app2 symbolic-ref HEAD) >actual &&
		test_cmp expect actual &&
		echo "module1: refs/heads/my/topic1" >expect &&
		(printf "module1: " && git -C projects/app1/module1 symbolic-ref HEAD) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "check tracking branch" '
	(
		cd work &&
		echo "main: refs/heads/maint" >expect &&
		(printf "main: " && git -C main config branch.my/topic1.merge) >actual &&
		test_cmp expect actual &&
		echo "driver1: refs/heads/maint" >expect &&
		(printf "driver1: " && git -C drivers/driver-1 config branch.my/topic1.merge) >actual &&
		test_cmp expect actual &&
		echo "driver2: refs/heads/maint" >expect &&
		(printf "driver2: " && git -C drivers/driver-2 config branch.my/topic1.merge) >actual &&
		test_cmp expect actual &&
		echo "app1: refs/heads/maint" >expect &&
		(printf "app1: " && git -C projects/app1 config branch.my/topic1.merge) >actual &&
		test_cmp expect actual &&
		echo "app2: refs/heads/maint" >expect &&
		(printf "app2: " && git -C projects/app2 config branch.my/topic1.merge) >actual &&
		test_cmp expect actual &&
		echo "module1: refs/heads/maint" >expect &&
		(printf "module1: " && git -C projects/app1/module1 config branch.my/topic1.merge) >actual &&
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
		my/topic2> Version 1.0.0
		EOF
		(
			printf "$(git -C projects/app1 symbolic-ref --short HEAD)> " &&
			git -C projects/app1 log -1 --pretty="%s"
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
			printf "$(git -C projects/app1 symbolic-ref --short HEAD)> " &&
			git -C projects/app1 log -1 --pretty="%s"
		) >actual &&
		test_cmp expect actual
	)
'

test_done
