#!/bin/sh

test_description="test 'git-repo sync' basic"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo sync (-n)" '
	(
		cd work &&
		git-repo init -u $manifest_url &&
		git-repo sync -n
	)
'

test_expect_success "git-repo sync (-n), no checkout" '
	(
		cd work &&
		test ! -d main &&
		test ! -d projects/app1 &&
		test ! -d projects/app2 &&
		test ! -d projects/app1/module1 &&
		test ! -d drivers/driver-1
	)
'


test_expect_success "git-repo sync (project-objects)" '
	(
		cd work/.repo/project-objects &&
		test -d drivers/driver1.git &&
		test ! -d drivers/driver2.git &&
		test -d project1.git &&
		test -d project2.git &&
		test -d project1/module1.git &&
		test -d main.git
	)
'

test_expect_success "git-repo sync (projects)" '
	(
		cd work/.repo/projects &&
		test -d drivers/driver-1.git &&
		test ! -d drivers/driver-2.git &&
		test -d projects/app1.git &&
		test -d projects/app2.git &&
		test -d projects/app1/module1.git &&
		test -d main.git
	)
'

test_expect_success "git-repo sync (size of project-objects)" '
	(
		cd work/.repo/project-objects &&
		test_size drivers/driver1.git -gt 100 &&
		test_size project1.git -gt 100 &&
		test_size project2.git -gt 100 &&
		test_size project1/module1.git -gt 100 &&
		test_size main.git -gt 100
	)
'

test_expect_success "git-repo sync (size of projects)" '
	(
		cd work/.repo/projects &&
		test_size drivers/driver-1.git -gt 100 &&
		test_size projects/app1.git -gt 100 &&
		test_size projects/app2.git -gt 100 &&
		test_size projects/app1/module1.git -gt 100 &&
		test_size main.git -gt 100
	)
'

test_expect_success "git-repo sync (-l)" '
	(
		cd work &&
		git-repo sync
	)
'

test_expect_success "git-repo sync (-l), checkouted" '
	(
		cd work &&
		test_size main -gt 15 &&
		test_size projects/app1 -gt 15 &&
		test_size projects/app2 -gt 15 &&
		test_size projects/app1/module1 -gt 15 &&
		test_size drivers/driver-1 -gt 15
	)
'

test_done
