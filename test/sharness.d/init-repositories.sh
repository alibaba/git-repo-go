#!/bin/sh

REPO_TEST_REPOSITORIES_VERSION=12

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

	# v0.1.0
	echo "# projecct: $name" >README.md &&
	echo v0.1.0 >VERSION &&
	printf "all:\n\t@echo \"$name: \$(shell cat VERSION)\"\n">Makefile &&
	git add README.md VERSION Makefile &&
	git commit -m "Version 0.1.0" &&
	git tag -m v0.1.0 v0.1.0 &&

	# v0.2.0
	echo v0.2.0 >VERSION &&
	git add -u &&
	git commit -m "Version 0.2.0" &&
	git tag -m v0.2.0 v0.2.0 &&

	# v0.3.0
	echo v0.3.0 >VERSION &&
	git add -u &&
	git commit -m "Version 0.3.0" &&
	git tag -m v0.3.0 v0.3.0 &&

	# v1.0.0
	echo v1.0.0 >VERSION &&
	git add -u &&
	git commit -m "Version 1.0.0" &&
	git tag -m v1.0.0 v1.0.0 &&

	# branch: master
	echo v2.0.0-dev >VERSION &&
	git add -u &&
	git commit -m "Version 2.0.0-dev" &&

	# refs/changes/45/12345/1
	git checkout v0.1.0 &&
	echo "$name: patch-1" >topic.txt &&
	git add topic.txt &&
	git commit -m "New topic" &&
	git update-ref refs/changes/45/12345/1 HEAD &&

	# refs/changes/45/12345/2
	echo "$name: patch-2" >topic.txt &&
	git add topic.txt &&
	git commit --amend -m "New topic" &&
	git update-ref refs/changes/45/12345/2 HEAD &&

	# refs/merge-requests/12345/head
	echo "$name: patch-3" >topic.txt &&
	git add topic.txt &&
	git commit --amend -m "New topic" &&
	git update-ref refs/merge-requests/12345/head HEAD &&

	# branch: Maint
	git checkout -b Maint v1.0.0 &&

	git push origin +refs/*:refs/* &&
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
	           revision="refs/heads/master"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	</manifest>
	<!-- base commit -->
	EOF
	git add default.xml &&
	git commit -m "base commit" &&

	# create tag v0.1
	git checkout master^{} &&
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
	<!-- tag v0.1 -->
	EOF
	git add default.xml &&
	git commit -m "Version 0.1" &&
	git tag -m v0.1 v0.1 &&

	# create tag v0.2
	git checkout master^{} &&
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
	<!-- tag v0.2 -->
	EOF
	git add default.xml &&
	git commit -m "Version 0.2" &&
	git tag -m v0.2 v0.2 &&

	# create tag v1.0 and branch Maint
	git checkout master^{} &&
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
	<!-- tag v1.0 -->
	EOF
	git add default.xml &&
	git commit -m "Version 1.0" &&
	git tag -m v1.0 v1.0 &&
	git branch Maint &&

	# update tag 2.0 and branch master
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
	<!-- tag v2.0 -->
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
	<!-- tag v2.0 -->
	EOF

	cat >remote-ro.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   revision="master"
		   review="https://example.com" />
	  <remote  name="others"
		   fetch=".."
		   revision="master" />
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
	  <project name="others/demo1" path="others/demo-1" groups="others" remote="others" />
	  <project name="others/demo2" path="others/demo-2" groups="notdefault,others" remote="others" />
	</manifest>
	<!-- tag v2.0 -->
	EOF

	git add default.xml next.xml remote-ro.xml &&
	git commit -m "Version 2.0" &&
	git tag -m v2.0 v2.0
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
		test_create_repository others/demo1.git &&
		test_create_repository others/demo2.git &&
		test_create_manifest_projects
	)
}

get_manifest_commits () {
	dir_m="$REPO_TEST_REPOSITORIES/hello/manifests.git"
	git -C $dir_m config core.abbrev 7 &&
	COMMIT_MANIFEST_MASTER=$(git -C $dir_m rev-parse master) &&
	COMMIT_MANIFEST_MAINT=$(git -C $dir_m rev-parse Maint) &&
	COMMIT_MANIFEST_0_1=$(git -C $dir_m rev-parse v0.1^0) &&
	COMMIT_MANIFEST_0_2=$(git -C $dir_m rev-parse v0.2^0) &&
	COMMIT_MANIFEST_1_0=$(git -C $dir_m rev-parse v1.0^0) &&
	COMMIT_MANIFEST_2_0=$(git -C $dir_m rev-parse v2.0^0) &&
	ABBREV_COMMIT_MANIFEST_MASTER=$(echo $COMMIT_MANIFEST_MASTER | cut -c 1-7) &&
	ABBREV_COMMIT_MANIFEST_MAINT=$(echo $COMMIT_MANIFEST_MAINT | cut -c 1-7) &&
	ABBREV_COMMIT_MANIFEST_0_1=$(echo $COMMIT_MANIFEST_0_1 | cut -c 1-7) &&
	ABBREV_COMMIT_MANIFEST_0_2=$(echo $COMMIT_MANIFEST_0_2 | cut -c 1-7) &&
	ABBREV_COMMIT_MANIFEST_1_0=$(echo $COMMIT_MANIFEST_1_0 | cut -c 1-7) &&
	ABBREV_COMMIT_MANIFEST_2_0=$(echo $COMMIT_MANIFEST_2_0 | cut -c 1-7)
}

# Run test_tick to initial author/committer name and time
test_tick

if ! test_repositories_is_uptodate
then
	repo_create_test_repositories
fi &&

get_manifest_commits
