#!/bin/sh

test_description="git-repo sync --detach"

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
			   review="https://code.aone.alibaba-inc.com" />
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

test_expect_success "git-repo sync" '
	(
		cd work &&
		git-repo init -u "$manifest_url" &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "projects are in detached mode" '
	(
		cd work &&
		(
			cd app1 &&
			test_must_fail git symbolic-ref HEAD &&
			cd ../app2 &&
			test_must_fail git symbolic-ref HEAD
		) >actual 2>&1 &&
		cat >expect <<-EOF &&
		fatal: ref HEAD is not a symbolic ref
		fatal: ref HEAD is not a symbolic ref
		EOF
		test_cmp expect actual
	)
'

test_expect_success "start to create new branch" '
	(
		cd work &&
		git repo start --all jx/topic &&
		(
			cd app1 &&
			git symbolic-ref HEAD &&
			cd ../app2 &&
			git symbolic-ref HEAD
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/jx/topic
		refs/heads/jx/topic
		EOF
		test_cmp expect actual
	)
'

test_expect_success "sync --detached" '
	(
		cd work &&
		git repo sync --detach \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		(
			cd app1 &&
			test_must_fail git symbolic-ref HEAD &&
			cd ../app2 &&
			test_must_fail git symbolic-ref HEAD
		) >actual 2>&1 &&
		cat >expect <<-EOF &&
		fatal: ref HEAD is not a symbolic ref
		fatal: ref HEAD is not a symbolic ref
		EOF
		test_cmp expect actual
	)
'

test_expect_success "switch branch" '
	(
		cd work &&
		git repo start --all jx/topic &&
		(
			cd app1 &&
			git symbolic-ref HEAD &&
			cd ../app2 &&
			git symbolic-ref HEAD
		) >actual &&
		cat >expect <<-EOF &&
		refs/heads/jx/topic
		refs/heads/jx/topic
		EOF
		test_cmp expect actual
	)
'

test_expect_success "edit and new commit" '
	(
		cd work &&
		git repo start --all jx/topic &&
		(
			cd app1 &&
			cat >VERSION <<-EOF &&
			app1: 2.0.1
			EOF
			cat >app1.txt <<-EOF &&
			app1: hack
			EOF
			git add app1.txt
		) &&
		(
			cd app2 &&
			cat >VERSION <<-EOF &&
			app2: 2.0.1
			EOF
			cat >app2.txt <<-EOF &&
			app2: hack
			EOF
			git add VERSION app2.txt &&
			test_tick &&
			git commit -m "app2: hack"
		) &&
		(
			cd app1 &&
			git log --pretty="%h %s" -1 &&
			cd ../app2 &&
			git log --pretty="%h %s" -1
		) >actual &&
		cat >expect <<-EOF &&
		9419d63 initial
		7cee347 app2: hack
		EOF
		test_cmp expect actual &&
		(
			cd app1 &&
			git status --porcelain
		) >actual &&
		cat >expect <<-EOF &&
		 M VERSION
		A  app1.txt
		EOF
		test_cmp expect actual
	)
'

test_expect_success "sync --detached" '
	(
		cd work &&
		git repo sync --detach \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		(
			cd app1 &&
			test_must_fail git symbolic-ref HEAD &&
			cd ../app2 &&
			test_must_fail git symbolic-ref HEAD
		) >actual 2>&1 &&
		cat >expect <<-EOF &&
		fatal: ref HEAD is not a symbolic ref
		fatal: ref HEAD is not a symbolic ref
		EOF
		test_cmp expect actual
	)
'

test_expect_success "edits are left in workspace" '
	(
		cd work &&
		(
			cd app1 &&
			git status --porcelain &&
			echo "----" &&
			cd ../app2 &&
			git status --porcelain
		) >actual &&
		cat >expect <<-EOF &&
		 M VERSION
		A  app1.txt
		----
		EOF
		test_cmp expect actual
	)
'

test_expect_success "new commit in app2" '
	(
		cd work &&
		(
			cd app2 &&
			cat >VERSION <<-EOF &&
			app2: 2.0.2
			EOF
			cat >app2.txt <<-EOF &&
			app2: hack2
			EOF
			git add VERSION app2.txt &&
			test_tick &&
			git commit -m "app2: hack2"
		) &&
		(
			cd app1 &&
			git log --pretty="%h %s" -1 &&
			cd ../app2 &&
			git log --pretty="%h %s" -1
		) >actual &&
		cat >expect <<-EOF &&
		9419d63 initial
		e2a965b app2: hack2
		EOF
		test_cmp expect actual &&
		(
			cd app1 &&
			git status --porcelain
		) >actual &&
		cat >expect <<-EOF &&
		 M VERSION
		A  app1.txt
		EOF
		test_cmp expect actual
	)
'

test_expect_success "sync --detached again" '
	(
		cd work &&
		git repo sync --detach \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		(
			cd app1 &&
			test_must_fail git symbolic-ref HEAD &&
			cd ../app2 &&
			test_must_fail git symbolic-ref HEAD
		) >actual 2>&1 &&
		cat >expect <<-EOF &&
		fatal: ref HEAD is not a symbolic ref
		fatal: ref HEAD is not a symbolic ref
		EOF
		test_cmp expect actual
	)
'

test_expect_success "edits are left in workspace" '
	(
		cd work &&
		(
			cd app1 &&
			git status --porcelain &&
			echo "----" &&
			cd ../app2 &&
			git status --porcelain
		) >actual &&
		cat >expect <<-EOF &&
		 M VERSION
		A  app1.txt
		----
		EOF
		test_cmp expect actual
	)
'

test_done
