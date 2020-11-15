#!/bin/sh

test_description="test 'git-repo list'"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -u $manifest_url -g app -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "git repo manifest: show manifest of Maint branch" '
	(
		cd work &&
		git-repo manifest
	) >actual &&
	cat >expect<<-EOF &&
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com"></remote>
	  <default remote="aone" revision="Maint" sync-j="4"></default>
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="refs/tags/v0.2.0" groups="app"></project>
	  <project name="project2" path="projects/app2" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_expect_success "git repo manifest: freeze manifest with partial sync" '
	(
		cd work &&
		git-repo manifest -r
	) 2>&1 | sed -e "s/\/.*\/trash directory.t1102-manifest-on-partial-sync/.../g" >actual &&
	cat >expect<<-EOF &&
	ERROR: cannot open git repo '"'"'.../work/.repo/projects/drivers/driver-1.git'"'"': repository does not exist
	WARNING: repository for drivers/driver1 is missing, fail to parse HEAD
	ERROR: cannot open git repo '"'"'.../work/.repo/projects/drivers/driver-2.git'"'"': repository does not exist
	WARNING: repository for drivers/driver2 is missing, fail to parse HEAD
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com"></remote>
	  <default remote="aone" revision="Maint" sync-j="4"></default>
	  <project name="main" path="main" revision="9bf4b931514c8eab528d41bc557949213a529846" groups="app" upstream="Maint">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" revision="a3946522edb40ee1693e879944ff35c7f379c608" groups="app" upstream="Maint"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="260da37cd2a35272375f0f3e64d917765b1d05e8" groups="app" upstream="refs/tags/v0.2.0"></project>
	  <project name="project2" path="projects/app2" revision="a256c3712bbe2bef657e64b3e8ac244b9e709dc4" groups="app" upstream="Maint"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_expect_success "git repo manifest: freeze manifest with partial sync, --suppress-upstream-revision" '
	(
		cd work &&
		git-repo manifest -r --suppress-upstream-revision
	) 2>&1 | sed -e "s/\/.*\/trash directory.t1102-manifest-on-partial-sync/.../g" >actual &&
	cat >expect<<-EOF &&
	ERROR: cannot open git repo '"'"'.../work/.repo/projects/drivers/driver-1.git'"'"': repository does not exist
	WARNING: repository for drivers/driver1 is missing, fail to parse HEAD
	ERROR: cannot open git repo '"'"'.../work/.repo/projects/drivers/driver-2.git'"'"': repository does not exist
	WARNING: repository for drivers/driver2 is missing, fail to parse HEAD
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com"></remote>
	  <default remote="aone" revision="Maint" sync-j="4"></default>
	  <project name="main" path="main" revision="9bf4b931514c8eab528d41bc557949213a529846" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" revision="a3946522edb40ee1693e879944ff35c7f379c608" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="260da37cd2a35272375f0f3e64d917765b1d05e8" groups="app"></project>
	  <project name="project2" path="projects/app2" revision="a256c3712bbe2bef657e64b3e8ac244b9e709dc4" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_done
