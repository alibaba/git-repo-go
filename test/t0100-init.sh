#!/bin/sh

test_description="git-repo init"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init -u" '
	(
		cd work &&
		git-repo init -u $manifest_url
	)
'

test_expect_success "check installed hooks" '
	cat >expect<<-EOF &&
	#!/bin/sh
	EOF
	head -1 .git-repo/hooks/commit-msg >actual &&
	test_cmp expect actual &&
	test -x .git-repo/hooks/commit-msg
'

test_expect_success "manifest points to default.xml" '
	(
		cd work &&
		test -f .repo/manifest.xml &&
		echo manifests/default.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "git config variable manifest.name = default.xml" '
	(
		cd work &&
		echo default.xml >expect &&
		(
			cd .repo/manifests &&
			git config manifest.name
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "two xml files checkout" '
	(
		cd work &&
		# Has three xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/next.xml
		.repo/manifests/remote-ro.xml
		EOF
		test_cmp expect actual
	)
'

test_expect_success "current branch = default" '
	(
		cd work &&
		echo "ref: refs/heads/default" >expect &&
		cp .repo/manifests.git/HEAD actual &&
		test_cmp expect actual
	)
'

test_expect_success "remote track: master" '
	(
		cd work &&
		cat >expect <<-EOF &&
		refs/heads/master
		EOF
		(
			cd .repo/manifests &&
			git config branch.default.merge
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "test init in subdir" '
	(
		cd work &&
		mkdir -p a/b/c &&
		cd a/b/c &&
		git-repo init -u $manifest_url &&
		test ! -d .repo &&
		cd "$HOME/work" &&
		rm -rf a
	)
'

test_expect_success "switch file: test init -m <file>" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master -m next.xml &&
		# manifest.xml => manifests/next.xml
		test -f .repo/manifest.xml &&
		echo manifests/next.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual &&
		# git config variable manifest.name => next.xml
		echo next.xml >expect &&
		(
			cd .repo/manifests &&
			git config manifest.name
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "switch branch: maint, no rollback" '
	(
		cd work &&
		git-repo init -u $manifest_url -b Maint -m default.xml &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		EOF
		ls .repo/manifests/*.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "after switch, remote track: maint" '
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

test_expect_success "no -b for repo init, use previous branch" '
	(
		cd work &&
		git-repo init -u $manifest_url &&
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

test_expect_success "detached manifest, drop default branch" '
	(
		cd work/.repo/manifests &&
		git checkout HEAD^0 &&
		git branch -D default
	)
'

test_expect_success "switch branch: Maint, file: default.xml" '
	(
		cd work &&
		git-repo init -u $manifest_url -b Maint -m default.xml &&
		# manifest.xml => manifests/default.xml
		echo manifests/default.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "check default branch and tracking branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			echo "----" &&
			git config branch.default.remote &&
			git config branch.default.merge &&
			git config manifest.name
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		----
		origin
		refs/heads/Maint
		default.xml
		EOF
		test_cmp expect actual
	)
'

test_expect_success "branch: Maint, no next.xml" '
	(
		cd work &&
		# Branch switched, no release.xml
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		EOF
		ls .repo/manifests/*.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "detached manifest, drop default branch" '
	(
		cd work/.repo/manifests &&
		git checkout HEAD^0 &&
		git branch -D default
	)
'

test_expect_success "switch to tag: v0.1" '
	(
		cd work &&
		git-repo init -u $manifest_url -b refs/tags/v0.1 &&
		(
			cd .repo/manifests &&
			test_must_fail git symbolic-ref HEAD &&
			test_must_fail git config branch.default.remote &&
			test_must_fail git config branch.default.merge &&
			git config manifest.name &&
			git describe
		) >actual &&
		cat >expect <<-EOF &&
		default.xml
		v0.1
		EOF
		test_cmp expect actual
	)
'

test_expect_success "switch branch: master, next.xml is back" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		# Has two xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/next.xml
		.repo/manifests/remote-ro.xml
		EOF
		test_cmp expect actual
	)
'

test_expect_success "check default branch and tracking branch" '
	(
		cd work &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			echo "----" &&
			git config branch.default.remote &&
			git config branch.default.merge &&
			git config manifest.name
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		----
		origin
		refs/heads/master
		default.xml
		EOF
		test_cmp expect actual
	)
'

test_done
