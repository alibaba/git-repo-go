#!/bin/sh

test_description="Test groups in git-repo init"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	(
		# create .repo file as a barrier, not find .repo deeper
		touch .repo &&
		mkdir work &&
		cd work &&
		git-repo init -u $manifest_url
	)
'

test_expect_success "no platform & group settings" '
	(
		cd work &&
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = all, groups = <default>" '
	(
		cd work &&
		git-repo init --platform all &&
		echo "default,platform-linux,platform-darwin,platform-windows" >expect &&
		git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = all, groups = all" '
	(
		cd work &&
		git-repo init --platform all --groups all &&
		echo "all,platform-linux,platform-darwin,platform-windows" >expect &&
		git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = auto, groups = <default>" '
	(
		cd work &&
		git-repo init --platform auto &&
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = auto, groups = default" '
	(
		cd work &&
		git-repo init --platform auto &&
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = <default>, groups = app" '
	(
		cd work &&
		git-repo init -g app &&
		echo "app,platform-*" >expect &&
		git -C .repo/manifests config manifest.groups | \
			sed -e "s/platform-[^,]*/platform-*/" \
			>actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = <default>, groups = <default> # nothing changed" '
	(
		cd work &&
		git-repo init -u $manifest_url
		echo "app,platform-*" >expect &&
		git -C .repo/manifests config manifest.groups | \
			sed -e "s/platform-[^,]*/platform-*/" \
			>actual &&
		test_cmp expect actual
	)
'

test_expect_success "platform = auto, groups = default" '
	(
		cd work &&
		git-repo init -p auto -g default -u $manifest_url
		printf "" >expect &&
		test_must_fail git -C .repo/manifests config manifest.groups >actual &&
		test_cmp expect actual
	)
'

test_done
