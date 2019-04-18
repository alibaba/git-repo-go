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

test_expect_success "git-repo init -b maint" '
	(
		cd work &&
		git-repo init -u $manifest_url -b maint
	)
'

test_expect_success "edit default.xml" '
	(
		cd work/.repo/manifests &&
		echo >>default.xml
	)
'

test_expect_success "git-repo init again without update" '
	(
		cd work &&
		git-repo init -u $manifest_url
	)
'

test_expect_success "git-repo init uppdate failed because of dirty" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: worktree of manifests is dirty, checkout failed
		EOF
		test_must_fail git-repo init -u $manifest_url -b master 2>actual &&
		test_cmp expect actual
	)
'

test_expect_success "update failed, remote track not changed" '
	(
		cd work &&
		cat >expect <<-EOF &&
		refs/heads/maint
		EOF
		git -C .repo/manifests config branch.default.merge >actual &&
		test_cmp expect actual
	)
'

test_expect_success "detached and remove default branch" '
	(
		cd work/.repo/manifests &&
		git checkout HEAD^0 &&
		git branch -D default
	)
'

test_expect_success "git-repo init ok for detached project even dirty" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		cat >expect <<-EOF &&
		refs/heads/master
		EOF
		git -C .repo/manifests config branch.default.merge >actual &&
		test_cmp expect actual
	)
'

test_done
