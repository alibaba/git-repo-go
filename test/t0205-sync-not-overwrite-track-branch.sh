#!/bin/sh

test_description="sync won't overwrite modified files in a tracking branch"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync to Maint branch" '
	(
		cd work &&
		git-repo init -u $manifest_url -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "create branch and track remote branch" '
	(
		cd work &&
		(cd drivers/driver-1 && git checkout -b test driver/Maint) &&
		(cd projects/app1 && git checkout -b test aone/Maint) &&
		(cd projects/app1/module1 && git checkout -b test aone/Maint) &&
		(cd projects/app2 && git checkout -b test aone/Maint)
	)
'

test_expect_success "edit files in workdir" '
	(
		cd work &&
		test -f drivers/driver-1/VERSION &&
		echo hacked >drivers/driver-1/VERSION &&
		test -f projects/app1/VERSION &&
		echo hacked >projects/app1/VERSION &&
		test -f projects/app1/module1/VERSION &&
		echo hacked >projects/app1/module1/VERSION &&
		test -f projects/app2/VERSION &&
		echo hacked >projects/app2/VERSION &&
		git -C projects/app2 add -A
	)
'

test_expect_success "fail to sync, workspace is dirty" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		test_must_fail git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests version: 2.0" '
	(
		cd work &&
		cat >expect<<-EOF &&
		manifests: Version 2.0
		EOF
		git -C .repo/manifests log -1 --pretty="manifests: %s" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "fail to sync drivers/driver-1, workspace is dirty (not staged)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		 M VERSION
		EOF
		git -C drivers/driver-1 status -uno --porcelain >actual &&
		test_cmp expect actual &&
		test_must_fail git-repo sync -l \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			-- drivers/driver-1 \
			2>&1 | grep "^Error:" >actual &&
		cat >expect <<-EOF &&
		Error: worktree of drivers/driver1 is dirty, checkout failed
		EOF
		test_cmp expect actual

	)
'

test_expect_success "fail to sync projects/app1, workspace is dirty (not staged)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		 M VERSION
		EOF
		git -C projects/app1 status -uno --porcelain >actual &&
		test_cmp expect actual &&
		test_must_fail git-repo sync -l \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			-- projects/app1 \
			2>&1 | grep "^Error:" >actual &&
		cat >expect <<-EOF &&
		Error: worktree of project1 is dirty, checkout failed
		EOF
		test_cmp expect actual

	)
'

test_expect_success "fail to sync projects/app2, workspace is dirty (staged)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		M  VERSION
		EOF
		git -C projects/app2 status -uno --porcelain >actual &&
		test_cmp expect actual &&
		test_must_fail git-repo sync -l \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"  \
			-- projects/app2 \
			2>&1 | grep "^Error:" >actual &&
		cat >expect <<-EOF &&
		Error: worktree of project2 is dirty, checkout failed
		EOF
		test_cmp expect actual

	)
'

test_expect_success "not overwrite modified files inside a tracking branch" '
	(
		cd work &&
		cat >expect <<-EOF &&
		drivers/driver-1/VERSION: hacked
		projects/app1/VERSION: hacked
		projects/app2/VERSION: hacked
		EOF
		echo "drivers/driver-1/VERSION: $(cat drivers/driver-1/VERSION)" >actual &&
		echo "projects/app1/VERSION: $(cat projects/app1/VERSION)" >>actual &&
		echo "projects/app2/VERSION: $(cat projects/app2/VERSION)" >>actual &&
		test_cmp expect actual
	)
'

test_done
