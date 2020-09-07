#!/bin/sh

test_description="check tracking branch for start command"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

git_repo_show_current_branch_and_tracking() {
	git-repo forall '
		echo "## $REPO_PATH" &&
		branch=$(git branch 2> /dev/null | sed -e "/^[^*]/d" -e "s/* \(.*\)/\1/") &&
		printf "   $branch => " &&
		(git config branch.${branch}.merge || true)
	'
}

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

test_expect_success "no tracking branch for app1/module1" '
	(
		cd work &&
		git-repo start --all jx/test1 &&
		git_repo_show_current_branch_and_tracking >actual &&
		cat >expect <<-EOF &&
		## main
		   jx/test1 => refs/heads/Maint
		## projects/app1
		   jx/test1 => refs/heads/Maint
		## projects/app1/module1
		   jx/test1 => 
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
		  <project name="main" path="main" revision="920edd5e44b7a62b01ce93314ad38521d8721278" groups="app" upstream="Maint">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" revision="9590ccd64309ee7cd5d97cd0c6ec52799e8e680d" groups="app" upstream="Maint"></project>
		  <project name="project1/module1" path="projects/app1/module1" revision="557abe6dbd27fabe9beda5570e563a428dc57765" groups="app" upstream="refs/tags/v0.2.0"></project>
		  <project name="project2" path="projects/app2" revision="8b32cf53a8d86812dc3f8557eb7628a4f5d4e27a" groups="app" upstream="Maint"></project>
		  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="69d4c0148193caf88bea68db0502e1021eebe8f1" groups="drivers" upstream="Maint"></project>
		  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers" upstream="Maint"></project>
		</manifest>
		EOF
		git add -u &&
		git commit -m "All projects except driver2 point to fixed revisions"
	)
'

test_expect_success "all projects except driver-2 have no tracking branch" '
	(
		cd work &&
		git-repo start --all jx/test2 &&
		git_repo_show_current_branch_and_tracking >actual &&
		cat >expect <<-EOF &&
		## main
		   jx/test2 => 
		## projects/app1
		   jx/test2 => 
		## projects/app1/module1
		   jx/test2 => 
		## projects/app2
		   jx/test2 => 
		## drivers/driver-1
		   jx/test2 => 
		## drivers/driver-2
		   jx/test2 => refs/heads/Maint
		EOF
		test_cmp expect actual
	)
'

test_done
