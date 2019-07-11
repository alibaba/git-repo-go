#!/bin/sh

REPO_TEST_REPOSITORIES_VERSION=9

# Create test repositories in .repositories
REPO_TEST_REPOSITORIES="${SHARNESS_TEST_SRCDIR}/test-repositories"
REPO_TEST_REPOSITORIES_VERSION_FILE="${REPO_TEST_REPOSITORIES}/.VERSION"

# Use fixed commit auther and committer
GIT_AUTHOR_EMAIL=author@example.com
GIT_AUTHOR_NAME='A U Thor'
GIT_COMMITTER_EMAIL=committer@example.com
GIT_COMMITTER_NAME='C O Mitter'
export GIT_AUTHOR_EMAIL GIT_AUTHOR_NAME
export GIT_COMMITTER_EMAIL GIT_COMMITTER_NAME

# Use fixed commit time
test_tick () {
	if test -z "${test_tick+set}"
	then
		test_tick=1112911993
	else
		test_tick=$(($test_tick + 60))
	fi
	GIT_COMMITTER_DATE="$test_tick -0700"
	GIT_AUTHOR_DATE="$test_tick -0700"
	export GIT_COMMITTER_DATE GIT_AUTHOR_DATE
}

repo_create_test_repositories () {
	# create lock
	lockmsg="locked by $$"
	while :
	do
		if test -f "${REPO_TEST_REPOSITORIES}.lock"
		then
			if test "$lockmsg" = "$(cat "${REPO_TEST_REPOSITORIES}.lock")"; then
				break
			fi
			echo >&2 "Another process is creating shared repositories: $(cat "${REPO_TEST_REPOSITORIES}.lock")"
			sleep 2

		else
			echo "$lockmsg" >"${REPO_TEST_REPOSITORIES}.lock"
		fi
	done

	if test_repositories_is_uptodate
	then
		return
	fi

	# Remove whole shared repositories
	echo >&2 "Will recreate shared repositories in $REPO_TEST_REPOSITORIES"
	rm -rf "$REPO_TEST_REPOSITORIES"

	# Start to create shared repositories
	repo_create_test_repositories_real

	# create version file
	echo ${REPO_TEST_REPOSITORIES_VERSION} >${REPO_TEST_REPOSITORIES_VERSION_FILE}

	# release the lock
	rm -f "${REPO_TEST_REPOSITORIES}.lock"
}

test_repositories_is_uptodate() {
	if test "$(cat "$REPO_TEST_REPOSITORIES_VERSION_FILE" 2>/dev/null)" = "${REPO_TEST_REPOSITORIES_VERSION}"
	then
		return 0
	fi
	return 1
}

test_create_repository () {
	repo=$1

	if test -z "$repo"
	then
		echo >&2 "Usage test_create_repository <reponame>"
		return 1
	fi

	if test "${repo%\.git}" = "$repo"
	then
		repo=${repo}.git
	fi

	name=$(basename $repo)
	name=${name%\.git}
	dir=$(dirname $repo)

	cd "$REPO_TEST_REPOSITORIES" &&
	if test -n "$dir" && test "$dir" != "."
	then
		mkdir -p "$dir"
	fi &&

	git init --bare "$repo" &&
	git clone "$repo" "tmp-$name" &&
	cd "tmp-$name" &&
	echo "# projecct: $name" >README.md &&
	echo v0.1.0 >VERSION &&
	printf "all:\n\t@echo \"$name: \$(shell cat VERSION)\"\n">Makefile &&
	git add README.md VERSION Makefile &&
	test_tick && git commit -m "Version 0.1.0" &&
	test_tick && git tag -m v0.1.0 v0.1.0 &&
	echo v0.2.0 >VERSION &&
	git add -u &&
	test_tick && git commit -m "Version 0.2.0" &&
	test_tick && git tag -m v0.2.0 v0.2.0 &&
	echo v0.3.0 >VERSION &&
	git add -u &&
	test_tick && git commit -m "Version 0.3.0" &&
	test_tick && git tag -m v0.3.0 v0.3.0 &&
	echo v1.0.0 >VERSION &&
	git add -u &&
	test_tick && git commit -m "Version 1.0.0" &&
	test_tick && git tag -m v1.0.0 v1.0.0 &&
	git branch Maint v1.0.0 &&
	echo v2.0.0-dev >VERSION &&
	git add -u &&
	test_tick && git commit -m "Version 2.0.0-dev" &&
	git push --tags origin master Maint &&
	git checkout v0.1.0 &&
	echo "$name: patch-1" >topic.txt &&
	git add topic.txt &&
	test_tick && git commit -m "New topic" &&
	git push origin HEAD:refs/changes/45/12345/1 &&
	echo "$name: patch-2" >topic.txt &&
	git add topic.txt &&
	test_tick && git commit --amend -m "New topic" &&
	git push origin HEAD:refs/changes/45/12345/2 &&
	echo "$name: patch-3" >topic.txt &&
	git add topic.txt &&
	test_tick && git commit --amend -m "New topic" &&
	git push origin HEAD:refs/merge-requests/12345/head &&
	cd "$REPO_TEST_REPOSITORIES" &&
	rm -rf "tmp-$name"
}

test_create_manifest_projects () {
	# create manifest repository
	cd "$REPO_TEST_REPOSITORIES" &&
	mkdir -p hello &&
	git init --bare hello/manifests.git &&
	git clone hello/manifests.git tmp-manifests &&
	cd tmp-manifests &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   review="https://example.com" />
	  <default remote="aone"
	           revision="refs/tags/v0.1.0"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	</manifest>
	EOF

	git add default.xml &&
	test_tick && git commit -m "Version 0.1" &&
	test_tick && git tag -m v0.1 v0.1 &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   review="https://example.com" />
	  <default remote="aone"
	           revision="refs/tags/v0.2.0"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app" revision="refs/tags/v0.1.0">
	    <project name="module1" path="module1" groups="app"/>
	  </project>
	</manifest>
	EOF

	git add default.xml &&
	test_tick && git commit -m "Version 0.2" &&
	test_tick && git tag -m v0.2 v0.2 &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   review="https://example.com" />
	  <remote  name="driver"
		   fetch=".."
		   review="https://example.com" />
	  <default remote="aone"
	           revision="Maint"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app">
	    <project name="module1" path="module1" groups="app" revision="refs/tags/v0.2.0" />
	  </project>
	  <project name="project2" path="projects/app2" groups="app"/>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
	</manifest>
	EOF

	git add default.xml &&
	test_tick && git commit -m "Version 1.0" &&
	test_tick && git tag -m v1.0 v1.0 &&
	git branch Maint &&

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
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app">
	    <project name="module1" path="module1" groups="app" revision="refs/tags/v1.0.0" />
	  </project>
	  <project name="project2" path="projects/app2" groups="app"/>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
	</manifest>
	EOF

	cat >next.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   revision="master"
		   review="https://example.com" />
	  <remote  name="driver"
		   fetch=".."
		   revision="master"
		   review="https://example.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app">
	    <project name="module1" path="module1" groups="notdefault,app"/>
	  </project>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
	  <project name="drivers/driver3" path="drivers/driver-3" groups="drivers" remote="driver" />
	</manifest>
	EOF

	git add default.xml next.xml &&
	test_tick && git commit -m "Version 2.0" &&
	test_tick && git tag -m v2.0 v2.0
	git push --tags origin Maint master &&

	cd "$REPO_TEST_REPOSITORIES" &&
	rm -rf "tmp-manifests"
}

repo_create_test_repositories_real () {
	mkdir -p "$REPO_TEST_REPOSITORIES" &&
	(
		test_create_repository hello/main.git &&
		test_create_repository hello/project1.git &&
		test_create_repository hello/project1/module1.git &&
		test_create_repository hello/project2.git &&
		test_create_repository drivers/driver1.git &&
		test_create_repository drivers/driver2.git &&
		test_create_repository drivers/driver3.git &&
		test_create_manifest_projects
	)
}

if ! test_repositories_is_uptodate
then
	repo_create_test_repositories
fi
