#!/bin/sh

REPO_TEST_REPOSITORIES_VERSION=3

# Create test repositories in .repositories
REPO_TEST_REPOSITORIES="${SHARNESS_TEST_SRCDIR}/test-repositories"
REPO_TEST_REPOSITORIES_VERSION_FILE="${REPO_TEST_REPOSITORIES}/.VERSION"

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
	echo v1.0 >VERSION &&
	echo "all:\n\t@echo \"$name: \$(shell cat VERSION)\"\n">Makefile &&
	git add README.md VERSION Makefile &&
	git commit -m "initial" &&
	git tag -m v1.0 v1.0 &&
	git branch maint &&
	echo v2.0-dev >VERSION &&
	git add VERSION &&
	git commit -m "bump version to 2.0-dev" &&
	git push origin master maint v1.0 &&
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
		   review="https://code.aone.alibaba-inc.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	</manifest>
	EOF

	git add default.xml &&
	git commit -m "Version 0.1" &&
	git tag -m v0.1 v0.1 &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   review="https://code.aone.alibaba-inc.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app">
	    <project name="module1" path="module1" groups="app"/>
	  </project>
	  <project name="project2" path="projects/app2" groups="app"/>
	</manifest>
	EOF

	git add default.xml &&
	git commit -m "Version 0.2" &&
	git tag -m v0.2 v0.2 &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   review="https://code.aone.alibaba-inc.com" />
	  <remote  name="driver"
		   fetch=".."
		   review="https://code.aone.alibaba-inc.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="main" path="main" groups="app">
	    <copyfile src="VERSION" dest="VERSION"></copyfile>
	    <linkfile src="Makefile" dest="Makefile"></linkfile>
	  </project>
	  <project name="project1" path="projects/app1" groups="app">
	    <project name="module1" path="module1" groups="app"/>
	  </project>
	  <project name="project2" path="projects/app2" groups="app"/>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="notdefault,drivers" remote="driver" />
	</manifest>
	EOF

	git add default.xml &&
	git commit -m "Version 1.0" &&
	git push -u origin master &&
	git tag -m v1.0 v1.0 &&
	git branch maint &&
	git push --tags origin maint master &&

	cat >next.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch="."
		   revision="maint"
		   review="https://code.aone.alibaba-inc.com" />
	  <remote  name="driver"
		   fetch=".."
		   revision="maint"
		   review="https://code.aone.alibaba-inc.com" />
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
	</manifest>
	EOF

	git add next.xml &&
	git commit -m "Version 2.0" &&
	git tag -m v2.0 v2.0
	git push --tags origin master &&

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
		test_create_manifest_projects
	)
}

if ! test_repositories_is_uptodate
then
	repo_create_test_repositories
fi
