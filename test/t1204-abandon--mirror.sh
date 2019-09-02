#!/bin/sh

test_description="git-repo abandon --mirror test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync" '
	(
		cd work &&
		git-repo init --mirror -u $manifest_url -g all -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			-- main
	)
'

test_expect_success "fail: cannot run abandon in mirrored repo" '
	(
		cd work/main.git &&
		test_must_fail git-repo abandon --all
	) >actual 2>&1 &&
	cat >expect <<-EOF &&
	FATAL: cannot run in a mirror
	EOF
	test_cmp expect actual
'

test_done
