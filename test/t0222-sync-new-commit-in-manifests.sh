#!/bin/sh

test_description="if manifests project changed, when running 'git-repo sync', manifests project should sync successful"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${HOME}/r/hello/manifests.git"

test_expect_success "setup" '
	cp -a "${REPO_TEST_REPOSITORIES}" r &&
	mkdir work
'

test_expect_success "init from default branch (master branch), and sync" '
	(
		cd work &&
		git-repo init -u "$manifest_url" &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "create new commit in manifest default branch" '
	git clone "$manifest_url" manifests &&
	(
		cd manifests &&
		git checkout master &&
		echo hello >master.txt &&
		git add master.txt &&
		test_tick &&
		git commit -m "manifest: update master branch" &&
		git push origin HEAD
	) &&
	COMMIT_TIP=$(git -C manifests rev-parse HEAD)
'

test_expect_success "sync again" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests updated successful" '
	(
		cd work/.repo/manifests &&
		git config branch.default.merge &&
		git status -s &&
		git log --pretty="%H %s" -1
	) >actual &&
	cat >expect <<-EOF &&
	refs/heads/master
	$COMMIT_TIP manifest: update master branch
	EOF
	test_cmp expect actual
'

test_expect_success "init -d and init -b Maint branch" '
	(
		cd work &&
		git-repo init -d &&
		git-repo init -b Maint
	) &&
	(
		cd work/.repo/manifests &&
		test ! -f master.txt &&
		git log -1 --pretty="%H %s"
	) >actual &&
	cat >expect <<-EOF &&
	$COMMIT_MANIFEST_MAINT Version 1.0
	EOF
	test_cmp expect actual
'

test_expect_success "create new commit in manifest Maint branch" '
	(
		cd manifests &&
		git checkout Maint &&
		echo world >maint.txt &&
		git add maint.txt &&
		test_tick &&
		git commit -m "manifest: update Maint branch" &&
		git push origin HEAD
	) &&
	COMMIT_TIP=$(git -C manifests rev-parse HEAD)
'

test_expect_success "sync again" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "manifests (track Maint branch) updated successful" '
	(
		cd work/.repo/manifests &&
		git config branch.default.merge &&
		git status -s &&
		test -f maint.txt &&
		test ! -f master.txt &&
		git log -1 --pretty="%H %s"
	) >actual &&
	cat >expect <<-EOF &&
	refs/heads/Maint
	$COMMIT_TIP manifest: update Maint branch
	EOF
	test_cmp expect actual
'

test_done
