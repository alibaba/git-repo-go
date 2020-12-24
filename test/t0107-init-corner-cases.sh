#!/bin/sh

test_description="git-repo init corner cases"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init -u -b refs/tags/v0.1" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.1 &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v0.1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "manifests project is detached" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			cat "$(git rev-parse --git-dir)/HEAD"
		) >actual &&
		cat >expect <<-EOF &&
		$COMMIT_MANIFEST_0_1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in manifests" '
	(
		cd work/.repo/manifests &&
		echo hello >>README.md &&
		git add README.md &&
		test_tick && git commit -m test
	) &&
	COMMIT_TIP=$(git -C work/.repo/manifests rev-parse HEAD)
'

test_expect_success "git-repo sync" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests project is still detached after sync" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			cat "$(git rev-parse --git-dir)/HEAD"
		) >actual &&
		cat >expect <<-EOF &&
		$COMMIT_TIP
		EOF
		test_cmp expect actual
	)
'

test_expect_success "git-repo init -u -b refs/tags/v0.2" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.2 &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v0.2
		EOF
		test_cmp expect actual
	)
'

test_expect_success "manifests project is detached after sync v0.2" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			cat "$(git rev-parse --git-dir)/HEAD"
		) >actual &&
		cat >expect <<-EOF &&
		$COMMIT_MANIFEST_0_2
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in manifests" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			echo hello >>README.md &&
			git add README.md &&
			test_tick && git commit -m test
		)
	)
'

test_expect_success "git-repo init -u -b master" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v2.0
		EOF
		test_cmp expect actual
	)
'

test_expect_success "manifests project checkout to default branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		refs/heads/master
		EOF
		test_cmp expect actual
	)
'

test_expect_success "git-repo sync" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests project still checkout to default branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		refs/heads/master
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in manifests" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			echo hello >>README.md &&
			git add README.md &&
			test_tick && git commit -m test
		)
	)
'

test_expect_success "git-repo init -u -b refs/tags/v0.1 failed" '
	(
		cd work &&
		test_must_fail git-repo init -u $manifest_url -b refs/tags/v0.1 >actual 2>&1 &&
		cat >expect <<-EOF &&
		Error: add --detach option to \`git repo init\` to throw away changes in '"'"'.repo/manifests'"'"'
		EOF
		test_cmp expect actual
	)
'

test_expect_success "remove new commit in manifests" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git reset --hard HEAD^
		)
	)
'

test_expect_success "git-repo init -u -b refs/tags/v0.1" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.1 &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v0.1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "still in default branch, but no tracking branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git describe --tags &&
			test_must_fail git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		v0.1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "git-repo sync" '
	(
		cd work &&
		test_must_fail git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	) 2>&1 | tail -6 >actual &&
	cat >expect <<-EOF &&
		ERROR: The following obsolete projects are still in your workspace, please check and remove them:
		 * drivers/driver-1
		 * projects/app1
		 * projects/app1/module1
		 * projects/app2
		Error: 4 obsolete projects in your workdir need to be removed
		EOF
	test_cmp expect actual
'

test_expect_success "after sync, manifest still in default branch, but no tracking branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git describe --tags &&
			test_must_fail git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		v0.1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in manifests" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			echo hello >>README.md &&
			git add README.md &&
			test_tick && git commit -m test
		)
	)
'

test_expect_success "git-repo init -u -b refs/heads/Maint" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/heads/Maint &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v1.0
		EOF
		test_cmp expect actual
	)
'

test_expect_success "tracking remote Maint branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		refs/heads/Maint
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in manifests" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			echo hello >>README.md &&
			git add README.md &&
			test_tick && git commit -m test
		)
	)
'

test_expect_success "git-repo init -u -b master failed" '
	(
		cd work &&
		test_must_fail git-repo init -u $manifest_url -b master >actual 2>&1 &&
		cat >expect <<-EOF &&
		Error: add --detach option to \`git repo init\` to throw away changes in '"'"'.repo/manifests'"'"'
		EOF
		test_cmp expect actual
	)
'

test_expect_success "Still tracking remote Maint branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		refs/heads/Maint
		EOF
		test_cmp expect actual
	)
'

test_expect_success "git-repo init -b master --detach" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master --detach &&
		(
			cd .repo/manifests &&
			git describe --tags
		) >actual &&
		cat >expect <<-EOF &&
		v2.0
		EOF
		test_cmp expect actual
	)
'

test_expect_success "manifests project is detached" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			cat "$(git rev-parse --git-dir)/HEAD"
		) >actual &&
		cat >expect <<-EOF &&
		$COMMIT_MANIFEST_MASTER
		EOF
		test_cmp expect actual
	)
'

test_done
