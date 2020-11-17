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
		git-repo init -u $manifest_url -g all --mirror -b Maint &&
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

test_expect_success "git repo manifest -o actual" '
	rm actual &&
	(
		cd work &&
		git-repo manifest -o ../actual
	) &&
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

test_expect_success "switch manifest branch" '
	(
		cd work &&
		git-repo init -b master &&
		git-repo sync
	)
'
	
test_expect_success "git repo manifest: show manifest of master branch" '
	(
		cd work &&
		git-repo manifest
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com" revision="Maint"></remote>
	  <default remote="aone" revision="master" sync-j="4"></default>
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="refs/tags/v1.0.0" groups="app"></project>
	  <project name="project2" path="projects/app2" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_expect_success "git repo manifest: freeze manifest" '
	(
		cd work &&
		git-repo manifest -r
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com" revision="Maint"></remote>
	  <default remote="aone" revision="master" sync-j="4"></default>
	  <project name="main" path="main" revision="4d13a6c1a2c17fcb3b109f2b1586d1485463e636" groups="app" upstream="master">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" revision="2fdfd9b9ff3bb556a74363bd0dacec0d29a0cc2a" groups="app" upstream="master"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="8fc882db0d6eaa24013f4ee3772e6765eb920d21" groups="app" upstream="refs/tags/v1.0.0"></project>
	  <project name="project2" path="projects/app2" revision="98dc74a3fac99714338633327dbab62b5189375b" groups="app" upstream="master"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="faa6f5cedc80d51cb57505376ef99878b66cd020" groups="drivers" upstream="Maint"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="df3d4c64f8d3be5365e1c778ba77976bda701c32" groups="notdefault,drivers" upstream="Maint"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_expect_success "git repo manifest: freeze manifest --suppress-upstream-revision" '
	(
		cd work &&
		git-repo manifest -r --suppress-upstream-revision
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	<manifest>
	  <remote name="aone" alias="origin" fetch="." review="https://example.com"></remote>
	  <remote name="driver" fetch=".." review="https://example.com" revision="Maint"></remote>
	  <default remote="aone" revision="master" sync-j="4"></default>
	  <project name="main" path="main" revision="4d13a6c1a2c17fcb3b109f2b1586d1485463e636" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" revision="2fdfd9b9ff3bb556a74363bd0dacec0d29a0cc2a" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" revision="8fc882db0d6eaa24013f4ee3772e6765eb920d21" groups="app"></project>
	  <project name="project2" path="projects/app2" revision="98dc74a3fac99714338633327dbab62b5189375b" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="faa6f5cedc80d51cb57505376ef99878b66cd020" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="df3d4c64f8d3be5365e1c778ba77976bda701c32" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_done
