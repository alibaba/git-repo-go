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
	  <project name="main" path="main" remote="aone" revision="Maint" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" remote="aone" revision="Maint" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" remote="aone" revision="refs/tags/v0.2.0" groups="app"></project>
	  <project name="project2" path="projects/app2" remote="aone" revision="Maint" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="Maint" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="Maint" groups="notdefault,drivers"></project>
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
	  <project name="main" path="main" remote="aone" revision="Maint" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" remote="aone" revision="Maint" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" remote="aone" revision="refs/tags/v0.2.0" groups="app"></project>
	  <project name="project2" path="projects/app2" remote="aone" revision="Maint" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="Maint" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="Maint" groups="notdefault,drivers"></project>
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
	  <project name="main" path="main" remote="aone" revision="master" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" remote="aone" revision="master" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" remote="aone" revision="refs/tags/v1.0.0" groups="app"></project>
	  <project name="project2" path="projects/app2" remote="aone" revision="master" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="master" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="master" groups="notdefault,drivers"></project>
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
	  <project name="main" path="main" remote="aone" revision="152dee6ad8698f9383cf8f6633a031ef3e99684a" groups="app" upstream="master">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" remote="aone" revision="eac322d281428f6c05d47ed96f7aafa12be956b4" groups="app" upstream="master"></project>
	  <project name="project1/module1" path="projects/app1/module1" remote="aone" revision="2be33cb731d4560f783341c74f0ec926d30377fd" groups="app" upstream="refs/tags/v1.0.0"></project>
	  <project name="project2" path="projects/app2" remote="aone" revision="927fd5d61f2edb798c226a0789c6da7c46a416f8" groups="app" upstream="master"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="f97e1371352714cbe4b686d292ce6c0579abbad2" groups="drivers" upstream="master"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="5ef3e8bc5d5299cdb245870124703cbf701b8599" groups="notdefault,drivers" upstream="master"></project>
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
	  <project name="main" path="main" remote="aone" revision="152dee6ad8698f9383cf8f6633a031ef3e99684a" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" remote="aone" revision="eac322d281428f6c05d47ed96f7aafa12be956b4" groups="app"></project>
	  <project name="project1/module1" path="projects/app1/module1" remote="aone" revision="2be33cb731d4560f783341c74f0ec926d30377fd" groups="app"></project>
	  <project name="project2" path="projects/app2" remote="aone" revision="927fd5d61f2edb798c226a0789c6da7c46a416f8" groups="app"></project>
	  <project name="drivers/driver1" path="drivers/driver-1" remote="driver" revision="f97e1371352714cbe4b686d292ce6c0579abbad2" groups="drivers"></project>
	  <project name="drivers/driver2" path="drivers/driver-2" remote="driver" revision="5ef3e8bc5d5299cdb245870124703cbf701b8599" groups="notdefault,drivers"></project>
	</manifest>
	EOF
	test_cmp expect actual
'

test_done
