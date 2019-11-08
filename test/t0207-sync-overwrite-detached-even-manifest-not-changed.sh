#!/bin/sh

test_description="sync overwrites modified files in detached head"

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
		git-repo init -u $manifest_url -b master &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "edit files in workdir, all projects are in detached HEAD" '
	(
		cd work &&
		test -f drivers/driver-1/VERSION &&
		echo hacked >drivers/driver-1/VERSION &&
		test -f projects/app1/VERSION &&
		echo hacked >projects/app1/VERSION &&
		test -f projects/app1/module1/VERSION &&
		echo hacked >projects/app1/module1/VERSION &&
		test -f projects/app2/VERSION &&
		echo hacked >projects/app2/VERSION
	)
'

test_expect_success "git-repo resync again" '
	(
		cd work &&
		git-repo init -u $manifest_url &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "nothing changed in manifests" '
	(
		cd work &&
		cat >expect<<-EOF &&
		manifests: Version 2.0
		EOF
		(
			cd .repo/manifests &&
			git log -1 --pretty="manifests: %s"
		) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "changes are preserved, even projects are in detached HEAD" '
	(
		cd work &&
		cat >expect<<-EOF &&
		driver-1: hacked
		app-1: hacked
		app-1.module1: hacked
		app-2: hacked
		EOF
		echo "driver-1: $(cat drivers/driver-1/VERSION)" >actual &&
		echo "app-1: $(cat projects/app1/VERSION)" >>actual &&
		echo "app-1.module1: $(cat projects/app1/module1/VERSION)" >>actual &&
		echo "app-2: $(cat projects/app2/VERSION)" >>actual &&
		test_cmp expect actual
	)
'

test_done
