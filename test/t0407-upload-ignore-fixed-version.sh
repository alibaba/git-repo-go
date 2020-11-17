#!/bin/sh

test_description="do not upload project with fixed revision"

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
		git-repo init -u $manifest_url -g all -b Maint
	)
'

test_expect_success "setup local manifest xml" '
	(
		cd work/.repo &&
		mkdir local_manifests &&

		cat >local_manifests/01-cleanup.xml <<-EOF &&
		<manifest>
		  <remove-project name="main" path="main"/>
		  <remove-project name="project1" path="projects/app1"/>
		  <remove-project name="project1/module1" path="projects/app1/module1"/>
		  <remove-project name="project2" path="projects/app2"/>
		  <remove-project name="drivers/driver1" path="drivers/driver-1"/>
		  <remove-project name="drivers/driver2" path="drivers/driver-2"/>
		</manifest>
		EOF

		cat >local_manifests/02-new-projects.xml <<-EOF &&
		<manifest>
		  <remote name="aone" alias="origin" fetch="." review="https://example.com" override="true"></remote>
		  <remote name="driver" fetch=".." review="https://example.com" override="true"></remote>
		  <default remote="aone" revision="Maint" sync-j="4" override="true"></default>

		  <project name="main" path="main" groups="app" revision="refs/tags/v1.0.0">
		    <copyfile src="VERSION" dest="VERSION"></copyfile>
		    <linkfile src="Makefile" dest="Makefile"></linkfile>
		  </project>
		  <project name="project1" path="projects/app1" groups="app" revision="master"></project>
		  <project name="project1/module1" path="projects/app1/module1" revision="refs/tags/v1.0.1" groups="app"></project>
		  <project name="project2" path="projects/app2" groups="app"></project>
		  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" groups="drivers"></project>
		  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers"></project>
		</manifest>
		EOF

		git repo manifest
	)
'

test_expect_success "git-repo init & sync" '
	(
		cd work &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "create local branch: my/topic" '
	(
		cd work &&
		git-repo start --all my/topic
	)
'

test_expect_success "check settings for branch tracking" '
	(
		cd work &&
		show_all_repo_branch_tracking >actual &&
		cat >expect-branch-tracking <<-EOF &&
		## main
		   my/topic => refs/heads/Maint
		## projects/app1
		   my/topic => refs/heads/master
		## projects/app1/module1
		   my/topic => refs/heads/Maint
		## projects/app2
		   my/topic => refs/heads/Maint
		## drivers/driver-1
		   my/topic => refs/heads/Maint
		## drivers/driver-2
		   my/topic => refs/heads/Maint
		EOF
		test_cmp expect-branch-tracking actual
	)
'

test_expect_failure "do not upload fixed revision which not changed" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF

		git-repo upload \
			--dryrun \
			--assume-yes \
			--no-cache \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"type\":\"agit\", \"version\":2, \"pushurl\":\"https://<email>@git.example.com/agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
			-e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" \
			-e "s#///.*/hello/main.git#///path/to/hello/main.git#g" \
			<out >actual &&
		test_cmp expect actual
	)
'

test_done
