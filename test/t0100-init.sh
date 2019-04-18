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
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual
	)
'

test_expect_success "two xml files checkout" '
	(
		cd work &&
		# Has two xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/next.xml
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
		git -C .repo/manifests config branch.default.merge >actual &&
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
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual
	)
'

test_expect_success "switch branch: maint, no rollback" '
	(
		cd work &&
		git-repo init -u $manifest_url -b maint &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/next.xml
		EOF
		ls .repo/manifests/*.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "after switch, remote track: maint" '
	(
		cd work &&
		cat >expect <<-EOF &&
		refs/heads/maint
		EOF
		git -C .repo/manifests config branch.default.merge >actual &&
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

test_expect_success "switch branch: maint, file: default.xml" '
	(
		cd work &&
		git-repo init -u $manifest_url -b maint -m default.xml &&
		# manifest.xml => manifests/default.xml
		echo manifests/default.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual
	)
'

test_expect_success "manifest.name => default.xml" '
	(
		cd work &&
		# git config variable manifest.name is set to default.xml
		echo default.xml >expect &&
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual
	)
'

test_expect_success "branch: maint, no next.xml" '
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

test_expect_success "switch branch: master, next.xml is back" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		# Has two xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/next.xml
		EOF
		test_cmp expect actual
	)
'

test_expect_success "again, remote track: master" '
	(
		cd work &&
		cat >expect <<-EOF &&
		refs/heads/master
		EOF
		git -C .repo/manifests config branch.default.merge >actual &&
		test_cmp expect actual
	)
'

test_done
