#!/bin/sh

test_description="Test groups in git-repo init"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${SHARED_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	(
		# create .repo file as a barrier, not find .repo deeper
		touch .repo &&
		mkdir work &&
		cd work &&
		git-repo init -u $manifest_url
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

test_done
