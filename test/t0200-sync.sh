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

test_expect_success "manifests version: 2.0" '
	(
		cd work &&
		cat >expect<<-EOF &&
		manifests: Version 2.0
		EOF
		git -C .repo/manifests log -1 --pretty="manifests: %s" >actual &&
		test_cmp expect actual
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

test_expect_success "check object repositories in .repo/project-objects" '
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

test_expect_success "check repositories in .repo/projects" '
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

test_expect_success "check size of .repo/project-objects" '
	(
		cd work/.repo/project-objects &&
		test_size drivers/driver1.git -gt 50 &&
		test_size project1.git -gt 50 &&
		test_size project2.git -gt 50 &&
		test_size project1/module1.git -gt 50 &&
		test_size main.git -gt 50
	)
'

test_expect_success "check size of .repo/projects" '
	(
		cd work/.repo/projects &&
		test_size drivers/driver-1.git -gt 50 &&
		test_size projects/app1.git -gt 50 &&
		test_size projects/app2.git -gt 50 &&
		test_size projects/app1/module1.git -gt 50 &&
		test_size main.git -gt 50
	)
'

test_expect_success "git-repo sync (-l)" '
	(
		cd work &&
		git-repo sync -l \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"

	)
'

test_expect_success "push.default is nothing after sync local-half" '
	(
		cd work &&
		git -C main config push.default &&
		git -C projects/app1 config push.default
		git -C projects/app1/module1 config push.default
		git -C projects/app2 config push.default
		git -C drivers/driver-1 config push.default
	) >actual &&
	cat >expect <<-EOF &&
	nothing
	nothing
	nothing
	nothing
	nothing
	EOF
	test_cmp expect actual
'

test_expect_success "no gerrit hooks for agit projects" '
	(
		cd work &&
		test -d .repo/project-objects/main.git/hooks &&
		test ! -f .repo/project-objects/main.git/hooks/commit-msg &&
		test -d .repo/project-objects/project1.git/hooks &&
		test ! -f .repo/project-objects/project1.git/hooks/commit-msg &&
		test -d .repo/project-objects/project2.git/hooks &&
		test ! -f .repo/project-objects/project2.git/hooks/commit-msg
	)
'

test_expect_success "git-repo sync (-l), checkouted" '
	(
		cd work &&
		test -f main/VERSION &&
		test -f projects/app1/VERSION &&
		test -f projects/app2/VERSION &&
		test -f projects/app1/module1/VERSION &&
		test -f drivers/driver-1/VERSION
	)
'

test_expect_success "check .repo/project.list" '
	(
		cd work &&
		test -f .repo/project.list &&
		cat >expect<<-EOF &&
		drivers/driver-1
		main
		projects/app1
		projects/app1/module1
		projects/app2
		EOF
		cp .repo/project.list actual &&
		test_cmp expect actual
	)
'

test_expect_success "copy and link files" '
	(
		cd work &&
		test -f .repo/project.list &&
		cat >expect<<-EOF &&
		main/Makefile
		EOF
		readlink Makefile >actual &&
		test_cmp expect actual &&
		test_cmp VERSION main/VERSION
	)
'

test_done
