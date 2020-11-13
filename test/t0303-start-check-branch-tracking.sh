#!/bin/sh

test_description="check tracking branch for start command"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init" '
	(
		cd work &&
		git-repo init -u $manifest_url -g all -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "app1/module1 also has tracking branch" '
	(
		cd work &&
		git-repo start --all jx/test1 &&
		show_all_repo_branch_tracking >actual &&
		cat >expect <<-EOF &&
		## main
		   jx/test1 => refs/heads/Maint
		## projects/app1
		   jx/test1 => refs/heads/Maint
		## projects/app1/module1
		   jx/test1 => refs/heads/Maint
		## projects/app2
		   jx/test1 => refs/heads/Maint
		## drivers/driver-1
		   jx/test1 => refs/heads/Maint
		## drivers/driver-2
		   jx/test1 => refs/heads/Maint
		EOF
		test_cmp expect actual
	)
'

test_expect_success "manifests: point to fixed revisions" '
	(
		cd work/.repo/manifests &&
		cat >default.xml <<-EOF &&
		<manifest>
		  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
		  <remote name="driver" fetch=".." review="https://example.com"></remote>
		  <default remote="aone" revision="Maint" sync-j="4"></default>
		  <project name="main" path="main" revision="920edd5e44b7a62b01ce93314ad38521d8721278" groups="app" upstream="A">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" revision="9590ccd64309ee7cd5d97cd0c6ec52799e8e680d" groups="app" dest-branch="B"></project>
		  <project name="project1/module1" path="projects/app1/module1" revision="557abe6dbd27fabe9beda5570e563a428dc57765" groups="app"></project>
		  <project name="project2" path="projects/app2" revision="8b32cf53a8d86812dc3f8557eb7628a4f5d4e27a" groups="app" upstream="refs/tags/v0.2.0"></project>
		  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="69d4c0148193caf88bea68db0502e1021eebe8f1" groups="drivers"></project>
		  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers" revision="C"></project>
		</manifest>
		EOF
		git add -u &&
		git commit -m "All projects except driver2 point to fixed revisions"
	)
'

test_expect_success "all projects have tracking branch" '
	(
		cd work &&
		git-repo start --all jx/test2 &&
		show_all_repo_branch_tracking >actual &&
		cat >expect <<-EOF &&
		## main
		   jx/test2 => refs/heads/A
		## projects/app1
		   jx/test2 => refs/heads/B
		## projects/app1/module1
		   jx/test2 => refs/heads/Maint
		## projects/app2
		   jx/test2 => refs/heads/Maint
		## drivers/driver-1
		   jx/test2 => refs/heads/Maint
		## drivers/driver-2
		   jx/test2 => refs/heads/C
		EOF
		test_cmp expect actual
	)
'

test_done
