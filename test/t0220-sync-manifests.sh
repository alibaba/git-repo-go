#!/bin/sh

test_description="sync will update and reload manifests project"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${HOME}/repositories/manifests.git"

test_expect_success "setup" '
	mkdir repositories &&
	git init --bare repositories/manifests.git &&
	git init --bare repositories/app1.git &&
	git init --bare repositories/app2.git &&
	(
		mkdir tmp &&
		cd tmp &&
		git clone --no-local ../repositories/manifests.git &&
		git clone --no-local ../repositories/app1.git &&
		git clone --no-local ../repositories/app2.git
	)
	touch .repo &&
	mkdir work
'

test_expect_success "setup repositories: manifests" '
	(
		cd tmp/manifests &&
		cat >default.xml <<-EOF &&
		<?xml version="1.0" encoding="UTF-8"?>
		<manifest>
		  <remote  name="origin"
			   fetch=".."
			   revision="master"
			   review="https://example.com" />
		  <default remote="origin"
			   revision="master"
			   sync-j="4" />
		  <project name="repositories/app1.git" path="app1" groups="app"/>
		  <project name="repositories/app2.git" path="app2" groups="app"/>
		</manifest>
		EOF
		git add default.xml &&
		test_tick &&
		git commit -m "initial" &&
		git push -u origin HEAD
	)
'

test_expect_success "setup repositories: app1" '
	(
		cd tmp/app1 &&
		cat >VERSION <<-EOF &&
		app1: 1.0.0
		EOF
		git add VERSION &&
		test_tick &&
		git commit -m "initial" &&
		git push -u origin HEAD
	)
'

test_expect_success "setup repositories: app2" '
	(
		cd tmp/app2 &&
		cat >VERSION <<-EOF &&
		app2: 1.0.0
		EOF
		git add VERSION &&
		test_tick &&
		git commit -m "initial" &&
		git push -u origin HEAD
	)
'

test_expect_success "git repo init" '
	(
		cd work &&
		git repo init -u "$manifest_url" &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.remote &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		origin
		refs/heads/master
		EOF
		test_cmp expect actual
	)
'

test_expect_success "git repo sync -d, manifests not detached" '
	(
		cd work &&
		git repo sync --detach \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.remote &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		origin
		refs/heads/master
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in remote manifests" '
	(
		cd tmp/manifests &&
		cat >remote.txt <<-EOF &&
		remote hack
		EOF
		git add remote.txt &&
		test_tick &&
		git commit -m "remote hack" &&
		git push -u origin HEAD
	)
'
test_expect_success "new commit in local manifests" '
	(
		cd work/.repo/manifests &&
		cat >local.txt <<-EOF &&
		local hack
		EOF
		test_tick &&
		git add local.txt &&
		git commit -m "local hack"
	)
'

test_expect_success "git repo sync -d, manifests not detached" '
	(
		cd work &&
		test_tick &&
		git repo sync --detach \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		(
			cd .repo/manifests &&
			git symbolic-ref HEAD &&
			git config branch.default.remote &&
			git config branch.default.merge
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/default
		origin
		refs/heads/master
		EOF
		test_cmp expect actual
	)
'

test_expect_success "local commit rebased" '
	(
		cd work/.repo/manifests &&
		git log --oneline
	) >actual &&
	cat >expect <<-EOF &&
	6edd9ab local hack
	a360f94 remote hack
	daa9359 initial
	EOF
	test_cmp expect actual
'

test_done
