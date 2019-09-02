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

test_expect_success "git-repo init --mirror -u" '
	(
		cd work &&
		git-repo init --mirror -u $manifest_url
	)
'

test_expect_success "git config repo.mirror" '
	(
		cd work/.repo/manifests.git &&
		git config -f config repo.mirror
	) >actual &&
	cat >expect <<-EOF &&
	true
	EOF
	test_cmp expect actual
'

test_done
