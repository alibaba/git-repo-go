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

test_expect_success "beforce sync local-half, gerrit hooks are not installed" '
	(
		cd work &&
		test -d .repo/project-objects/main.git/hooks &&
		test -d .repo/project-objects/project1.git/hooks &&
		test -d .repo/project-objects/project2.git/hooks &&
		test -d .repo/project-objects/project1/module1.git/hooks &&
		test -d .repo/project-objects/drivers/driver1.git/hooks &&
		test ! -e .repo/project-objects/main.git/hooks/commit-msg &&
		test ! -e .repo/project-objects/project1.git/hooks/commit-msg &&
		test ! -e .repo/project-objects/project2.git/hooks/commit-msg &&
		test ! -e .repo/project-objects/project1/module1.git/hooks/commit-msg &&
		test ! -e .repo/project-objects/drivers/driver1.git/hooks/commit-msg
	)
'

test_expect_success "git-repo sync (-l), server has gerrit response" '
	(
		cd work &&
		git-repo sync -l \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"ssh.example.com 29418"

	)
'

test_expect_success "push.default is nothing after sync local-half" '
	(
		cd work &&
		( cd main && git config push.default) &&
		( cd projects/app1 && git config push.default ) &&
		( cd projects/app1/module1 && git config push.default ) &&
		( cd projects/app2 && git config push.default ) &&
		( cd drivers/driver-1 && git config push.default )
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

test_expect_success "projects hooks link to project-objects hooks" '
	(
		cd work &&
		test -L .repo/projects/main.git/hooks &&
		test -L .repo/projects/projects/app1.git/hooks &&
		test -L .repo/projects/projects/app2.git/hooks &&
		test -L .repo/projects/projects/app1/module1.git/hooks &&
		test -L .repo/projects/drivers/driver-1.git/hooks
	)
'

test_expect_success "Installed gerrit hooks for gerrit projects" '
	(
		cd work &&
		test -d .repo/project-objects/main.git/hooks &&
		test -d .repo/project-objects/project1.git/hooks &&
		test -d .repo/project-objects/project2.git/hooks &&
		test -d .repo/project-objects/project1/module1.git/hooks &&
		test -d .repo/project-objects/drivers/driver1.git/hooks &&
		test -L .repo/project-objects/main.git/hooks/commit-msg &&
		test -L .repo/project-objects/project1.git/hooks/commit-msg &&
		test -L .repo/project-objects/project2.git/hooks/commit-msg &&
		test -L .repo/project-objects/project1/module1.git/hooks/commit-msg &&
		test -L .repo/project-objects/drivers/driver1.git/hooks/commit-msg
	)
'

test_done
