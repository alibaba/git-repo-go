#!/bin/sh

test_description="sync with rebased/squashed commit"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${HOME}/r/hello/manifests.git"

test_expect_success "setup" '
	cp -a "${REPO_TEST_REPOSITORIES}" r &&
	mkdir work
'

test_expect_success "init from Maint branch and sync" '
	(
		cd work &&
		git-repo init -u "$manifest_url" -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "create a squash commit in projects/app1" '
	(
		cd work &&
		git repo start --all my/topic &&
		cd projects/app1 &&
		echo topic1 >topic1.txt &&
		echo topic2 >topic2.txt &&
		git add topic1.txt topic2.txt &&
		test_tick &&
		git commit -m "app1: squash topic1 & topic2" &&
		git push aone HEAD:Maint
	)
'

test_expect_success "recreate local commits" '
	(
		cd work &&
		cd projects/app1 &&
		git rm topic2.txt &&
		test_tick &&
		git commit --amend -m "app1: add topic1" &&
		echo topic3 >topic3.txt &&
		git add topic3.txt &&
		test_tick &&
		git commit -m "app1: add topic3"
	)
'

test_expect_success "sync network-only, and show commit log" '
	(
		cd work &&
		git-repo sync -n \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	) &&
	(
		cd work/projects/app1 &&
		git log --oneline -3
	) >actual &&
	cat >expect <<-EOF &&
	b840c85 app1: add topic3
	d403463 app1: add topic1
	c8e033e Version 1.0.0
	EOF
	test_cmp expect actual &&
	(
		cd work/projects/app1 &&
		git log --oneline -2 aone/Maint
	) >actual &&
	cat >expect <<-EOF &&
	c1dc349 app1: squash topic1 & topic2
	c8e033e Version 1.0.0
	EOF
	test_cmp expect actual
'

test_expect_success "rebased after sync" '
	(
		cd work &&
		test_tick &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	) &&
	(
		cd work/projects/app1 &&
		git log --oneline -3
	) >actual &&
	cat >expect <<-EOF &&
	4815294 app1: add topic3
	c1dc349 app1: squash topic1 & topic2
	c8e033e Version 1.0.0
	EOF
	test_cmp expect actual
'

test_done
