#!/bin/sh

test_description="sync with special manifests file"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${HOME}/repositories/hello/manifests"

test_expect_success "setup" '
	(
		# create .repo file as a barrier, not find .repo deeper
		touch .repo &&
		mkdir repositories &&
		cd repositories &&
		ln -s "${REPO_TEST_REPOSITORIES}/drivers" . &&
		ln -s "${REPO_TEST_REPOSITORIES}/others"  . &&
		mkdir hello &&
		cd hello &&
		ln -s "${REPO_TEST_REPOSITORIES}/hello/main.git" . &&
		ln -s "${REPO_TEST_REPOSITORIES}/hello/project1.git" . &&
		ln -s "${REPO_TEST_REPOSITORIES}/hello/project2.git" . &&
		ln -s "${REPO_TEST_REPOSITORIES}/hello/project1" . &&
		git clone --mirror \
			"${REPO_TEST_REPOSITORIES}/hello/manifests.git" \
			manifests.git
	)
'

test_expect_success "setup manifests: no name attr in project element" '
	(
		git clone "${manifest_url}.git" manifests &&
		cd manifests &&
		git checkout master &&

		cat >default.xml <<-EOF &&
		<?xml version="1.0" encoding="UTF-8"?>
		<manifest>
		  <remote  name="aone"
			   alias="origin"
			   fetch="."
			   review="https://example.com" />
		  <remote  name="driver"
			   fetch=".."
			   revision="Maint"
			   review="https://example.com" />
		  <default remote="aone"
			   revision="master"
			   sync-j="4" />
		  <project name="main" groups="app">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" groups="app">
		    <project name="module1" groups="app" revision="refs/tags/v1.0.0" />
		  </project>
		  <project path="projects/app2" groups="app"/>
		  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
		  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
		</manifest>
		<!-- no name attr in project element -->
		EOF
		git add -u &&
		git commit -m "default.xml: no name attr in project element" &&
		git push
	)
'

test_expect_success "fail to sync: empty name attr in project" '
	(
		mkdir work1 &&
		cd work1 &&
		git-repo init -u "$manifest_url" &&
		test_must_fail git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" >out 2>&1 &&
		grep "^FATAL" out >actual &&
		cat >expect <<-\EOF &&
		FATAL: "project" element has empty "name"
		EOF
		test_cmp expect actual
	)
'


test_expect_success "setup manifests: no name attr in remote element" '
	(
		cd manifests &&

		cat >default.xml <<-EOF &&
		<?xml version="1.0" encoding="UTF-8"?>
		<manifest>
		  <remote 
			   alias="origin"
			   fetch="."
			   review="https://example.com" />
		  <remote 
			   fetch=".."
			   revision="Maint"
			   review="https://example.com" />
		  <default remote="aone"
			   revision="master"
			   sync-j="4" />
		  <project name="main" groups="app">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" groups="app">
		    <project name="module1" groups="app" revision="refs/tags/v1.0.0" />
		  </project>
		  <project name="projects/app2" path="projects/app2" groups="app"/>
		  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
		  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
		</manifest>
		<!-- no name attr in remote element -->
		EOF
		git add -u &&
		git commit -m "default.xml: no name attr in remote element" &&
		git push
	)
'

test_expect_success "fail to sync: empty name attr in remote element" '
	(
		mkdir work2 &&
		cd work2 &&
		git-repo init -u "$manifest_url" &&
		test_must_fail git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" >out 2>&1 &&
		grep "^FATAL" out >actual &&
		cat >expect <<-\EOF &&
		FATAL: "remote" element has empty "name"
		EOF
		test_cmp expect actual
	)
'

test_expect_success "setup manifests: no path attr in project element" '
	(
		cd manifests &&

		cat >default.xml <<-EOF &&
		<?xml version="1.0" encoding="UTF-8"?>
		<manifest>
		  <remote  name="aone"
			   alias="origin"
			   fetch="."
			   review="https://example.com" />
		  <remote  name="driver"
			   fetch=".."
			   revision="Maint"
			   review="https://example.com" />
		  <default remote="aone"
			   revision="master"
			   sync-j="4" />
		  <project name="main" groups="app">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" groups="app">
		    <project name="module1" groups="app" revision="refs/tags/v1.0.0" />
		  </project>
		  <project name="project2" path="projects/app2" groups="app"/>
		  <project name="drivers/driver1" groups="drivers" remote="driver" />
		  <project name="drivers/driver2" groups="notdefault,drivers" remote="driver" />
		</manifest>
		<!-- no path attr in project element -->
		EOF
		git add -u &&
		git commit -m "default.xml: no path attributes for some projects" &&
		git push
	)
'

test_expect_success "sync ok for project without path attr" '
	(
		mkdir work3 &&
		cd work3 &&
		git-repo init -u "$manifest_url" &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		test -f main/VERSION &&
		test -f projects/app1/VERSION &&
		test -f projects/app2/VERSION &&
		test -f projects/app1/module1/VERSION &&
		test -f drivers/driver1/VERSION
	)
'

test_done
