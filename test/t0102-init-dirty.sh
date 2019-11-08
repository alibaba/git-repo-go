#!/bin/sh

test_description="git-repo init with dirty worktree"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init -b Maint" '
	(
		cd work &&
		git-repo init -u $manifest_url -b Maint
	)
'

test_expect_success "edit default.xml" '
	(
		cd work/.repo/manifests &&
		test -f default.xml &&
		echo >>default.xml
	)
'

test_expect_success "no upstream changed, init ok" '
	(
		cd work &&
		git-repo init -u $manifest_url
	)
'

test_expect_success "init -b to change branch, failed for dirty" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: worktree of manifests is dirty, checkout failed
		EOF
		test_must_fail git-repo init -u $manifest_url -b master 2>actual &&
		test_cmp expect actual
	)
'

test_expect_success "no change for remote track" '
	(
		cd work &&
		cat >expect <<-EOF &&
		refs/heads/Maint
		EOF
		(
			cd .repo/manifests &&
			git config branch.default.merge
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "detached by hand" '
	(
		cd work/.repo/manifests &&
		git checkout HEAD^0 &&
		git branch -D default
	)
'

test_expect_success "do not overwrite dirty file" '
	(
		cd work &&
		test_must_fail git-repo init -u $manifest_url -b master 2>&1 |head -2 >actual &&
		cat >expect <<-EOF &&
		error: Your local changes to the following files would be overwritten by checkout:
		EOF
		printf "\tdefault.xml\n" >>expect &&
		test_cmp expect actual
	)
'

test_expect_success "clean dirty manifest repo" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git checkout -- . &&
			git status --porcelain
		) >actual &&
		cat >expect <<-EOF &&
		EOF
		test_cmp expect actual
	)
'

test_expect_success "detached using init --detach command" '
	(
		cd work &&
		git-repo init --detach
	)
'

test_expect_success "touble detach" '
	(
		cd work &&
		git-repo init --detach
	)
'

test_expect_success "init switched to a tag" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.1
	)
'

test_expect_success "tag points to version 0.1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Version 0.1
		EOF
		(
			cd .repo/manifests &&
			git log -1 --pretty="%s"
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "has one xml file" '
	(
		cd work &&
		# Has two xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		EOF
		test_cmp expect actual
	)
'

test_expect_success "edit default.xml" '
	(
		cd work/.repo/manifests &&
		test -f default.xml &&
		echo >>default.xml
	)
'

test_expect_success "fail to switch, because of dirty manifest" '
	(
		cd work &&
		test_must_fail git-repo init -u $manifest_url -b Maint 2>&1 |head -3 >actual &&
		cat >expect <<-EOF &&
		NOTE: manifests> leaving default; does not track upstream
		error: Your local changes to the following files would be overwritten by checkout:
		EOF
		printf "\tdefault.xml\n" >>expect &&
		test_cmp expect actual
	)
'

test_done
