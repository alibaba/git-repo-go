#!/bin/sh

test_description="sync overwrites modified files in branch without a remote tracking branch"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync to maint branch" '
	(
		cd work &&
		git-repo init -u $manifest_url -b maint &&
		git-repo sync
	)
'

test_expect_success "create branch, but do not track remote branch" '
	(
		cd work &&
		(cd drivers/driver-1 && git checkout -b test) &&
		(cd projects/app1 && git checkout -b test) &&
		(cd projects/app1/module1 && git checkout -b test) &&
		(cd projects/app2 && git checkout -b test)
	)
'

test_expect_success "edit files in workdir" '
	(
		cd work &&
		test -f drivers/driver-1/VERSION &&
		echo hacked >drivers/driver-1/VERSION &&
		test -f projects/app1/VERSION &&
		echo hacked >projects/app1/VERSION &&
		test -f projects/app1/module1/VERSION &&
		echo hacked >projects/app1/module1/VERSION &&
		test -f projects/app2/VERSION &&
		echo hacked >projects/app2/VERSION &&
		git -C projects/app2 add -A
	)
'

test_expect_success "current branch which not has a remote track branch, will be overwritten" '
	(
		cd work &&
		git-repo init -u $manifest_url -b master &&
		git-repo sync &&
		cat >expect <<-EOF &&
		drivers/driver-1/VERSION: v2.0.0-dev
		projects/app1/VERSION: v2.0.0-dev
		projects/app2/VERSION: v2.0.0-dev
		EOF
		echo "drivers/driver-1/VERSION: $(cat drivers/driver-1/VERSION)" >actual &&
		echo "projects/app1/VERSION: $(cat projects/app1/VERSION)" >>actual &&
		echo "projects/app2/VERSION: $(cat projects/app2/VERSION)" >>actual &&
		test_cmp expect actual
	)
'

test_done
