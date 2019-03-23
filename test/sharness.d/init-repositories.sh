#!/bin/sh

SHARED_REPOSITORIES_VERSION=0

# Create test repositories in .repositories
SHARED_REPOSITORIES="${SHARNESS_TEST_SRCDIR}/test-repositories"
SHARED_REPOSITORIES_VERSION_FILE="${SHARED_REPOSITORIES}/.VERSION"

test_create_shared_repositories () {
	# create lock
	lockmsg="locked by $$"
	while :
	do
		if test -f "${SHARED_REPOSITORIES}.lock"
		then
			if test "$lockmsg" = "$(cat "${SHARED_REPOSITORIES}.lock")"; then
				break
			fi
			echo >&2 "Another process is creating shared repositories: $(cat "${SHARED_REPOSITORIES}.lock")"
			sleep 2

		else
			echo "$lockmsg" >"${SHARED_REPOSITORIES}.lock"
		fi
	done

	if test_shared_repositories_version
	then
		return
	fi

	# Remove whole shared repositories
	echo >&2 "Will recreate shared repositories in $SHARED_REPOSITORIES"
	rm -rf "$SHARED_REPOSITORIES"

	# Start to create shared repositories
	test_create_shared_repositories_real

	# create version file
	echo ${SHARED_REPOSITORIES_VERSION} >${SHARED_REPOSITORIES_VERSION_FILE}

	# release the lock
	rm -f "${SHARED_REPOSITORIES}.lock"
}

test_shared_repositories_version() {
	if test "$(cat "$SHARED_REPOSITORIES_VERSION_FILE" 2>/dev/null)" = "${SHARED_REPOSITORIES_VERSION}"
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

	cd "$SHARED_REPOSITORIES" &&
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
	git tag v1.0 &&
	git branch maint &&
	echo v2.0-dev >VERSION &&
	git add VERSION &&
	git commit -m "bump version to 2.0-dev" &&
	git push origin master maint v1.0 &&
	cd "$SHARED_REPOSITORIES" &&
	rm -rf "tmp-$name"
}

test_create_manifest_projects () {
	# create manifest repository
	cd "$SHARED_REPOSITORIES" &&
	mkdir -p hello &&
	git init --bare hello/manifests.git &&
	git clone hello/manifests.git tmp-manifests &&
	cd tmp-manifests &&

	cat >default.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch=".."
		   review="https://code.aone.alibaba-inc.com" />
	  <remote  name="driver"
		   fetch="../.."
		   review="https://code.aone.alibaba-inc.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="hello/main" path="main" groups="app">
	    <copyfile src="VERSION" dest="../VERSION"></copyfile>
	    <linkfile src="Makefile" dest="../Makefile"></linkfile>
	  </project>
	  <project name="hello/project1" path="projects/app1" groups="app">
	    <project name="hello/project1/module1" path="module1" groups="app"/>
	  </project>
	  <project name="hello/project2" path="projects/app2" groups="app"/>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="drivers" remote="driver" />
	</manifest>
	EOF

	git add default.xml &&
	git commit -m "initial" &&
	git push -u origin master &&
	git tag -m v1.0 v1.0 &&
	git branch maint &&
	git push origin v1.0 maint master &&

	cat >release.xml <<-EOF &&
	<?xml version="1.0" encoding="UTF-8"?>
	<manifest>
	  <remote  name="aone"
	           alias="origin"
		   fetch=".."
		   revision="maint"
		   review="https://code.aone.alibaba-inc.com" />
	  <remote  name="driver"
		   fetch="../.."
		   revision="maint"
		   review="https://code.aone.alibaba-inc.com" />
	  <default remote="aone"
	           revision="master"
		   sync-j="4" />
	  <project name="hello/main" path="main" groups="app">
	    <copyfile src="VERSION" dest="../VERSION"></copyfile>
	    <linkfile src="Makefile" dest="../Makefile"></linkfile>
	  </project>
	  <project name="hello/project1" path="projects/app1" groups="app">
	    <project name="hello/project1/module1" path="module1" groups="app"/>
	  </project>
	  <project name="drivers/driver1" path="drivers/driver-1" groups="drivers" remote="driver" />
	  <project name="drivers/driver2" path="drivers/driver-2" groups="drivers" remote="driver" />
	</manifest>
	EOF

	git add release.xml &&
	git commit -m "Add new xml: releaes.xml" &&
	git push origin master &&

	cd "$SHARED_REPOSITORIES" &&
	rm -rf "tmp-manifests"
}

test_create_shared_repositories_real () {
	mkdir -p "$SHARED_REPOSITORIES" &&
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

if ! test_shared_repositories_version
then
	test_create_shared_repositories
fi
