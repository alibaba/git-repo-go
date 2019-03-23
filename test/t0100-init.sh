#!/bin/sh

test_description="git-repo init"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${SHARED_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init" '
	(
		cd work &&
		git-repo init -u $manifest_url &&
		# after init, manifest.xml links to manifests/default.xml
		test -f .repo/manifest.xml &&
		echo manifests/default.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual &&
		# git config variable manifest.name is set to default.xml
		echo default.xml >expect &&
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual &&
		# Has two xml files
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		.repo/manifests/release.xml
		EOF
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

test_expect_success "test init -m <file>" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master -m release.xml &&
		# after init, manifest.xml links to manifests/release.xml
		test -f .repo/manifest.xml &&
		echo manifests/release.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual &&
		# git config variable manifest.name is set to default.xml
		echo release.xml >expect &&
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual
	)
'

test_expect_success "group test" '
	(
		cd work &&
		# platform = auto
		git-repo init --platform auto &&
		echo platform- >expect &&
		git -C .repo/manifests config manifest.groups | sed -e "s/-.*$/-/g">actual &&
		test_cmp expect actual &&
		# platform = all
		git-repo init --platform all &&
		echo "platform-linux,platform-darwin,platform-windows" >expect &&
		git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual &&
		# platform = auto, group = default
		git-repo init -g default --platform auto &&
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual &&
		# -g app
		git-repo init -g app &&
		echo app >expect &&
		git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual &&
		# usee default platform and group
		git-repo init &&
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual &&
		# -g app -p linux
		git-repo init -g app -p linux &&
		echo app,platform-linux >expect &&
		git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "switch branch" '
	(
		cd work &&
		git-repo init -u $manifest_url -b maint -m default.xml &&
		# after init, manifest.xml links to manifests/default.xml
		echo manifests/default.xml >expect &&
		readlink .repo/manifest.xml >actual &&
		test_cmp expect actual &&
		# git config variable manifest.name is set to default.xml
		echo default.xml >expect &&
		git -C .repo/manifests config manifest.name >actual &&
		test_cmp expect actual &&
		# Branch switched, no release.xml
		ls .repo/manifests/*.xml >actual &&
		cat >expect<<-EOF &&
		.repo/manifests/default.xml
		EOF
		test_cmp expect actual
	)
'

test_done
