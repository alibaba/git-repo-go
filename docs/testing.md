# Unit test

Unit test cases are written in golang test framework, see files with filename
end with `_test.go`.


# Integration tests in shell scripts (sharness test framework)

It will be too complecated to write integration tests using golang. We know
that `git.git` project using shell scripts for testing.  How about writing
test cases in shell scirpt?

[Sharness](https://github.com/chriscool/sharness) is a open source project,
which can help to write integration test cases in shell script. It is based
on git.git test framework created by Junio, can be used in any program.

You can find test cases written in shell scripts in `test/` directory.
As how to write and run test cases, please visit:

* https://github.com/chriscool/sharness
* https://github.com/git/git/blob/master/t/README


# Create sample repositories

Shell scripts in `test/sharness.d/` will be run before any test script, so we
can define global functions by creating shell scripts in it.

In order to test manifest project and projects defined by manifest XML, we
need to create many repositories as templates for test cases.  Shell script
` test/sharness.d/init-repositories.sh` is created for this.

If you want to change the manifest file, update the `test_create_manifest_projects`
function.

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
    
        ... ...
    }

If the layout of the repositories, contents of manifest file, and others are
changed, you must increase the version number below, so that the example
repositories can be recreated.

    SHARED_REPOSITORIES_VERSION=0


# Create new test cases

To write new test suite in file `t<NNNN>-<name>.sh` in `test/` directory,
you can see files, such as `t0001-test-log.sh`, `t0100-init.sh` for reference.

The two digits of the four numbers of the filename is used to group similar
test suites.  For example, test suites for `git-repo init` start with `t01`,
such as `t0100-init.sh`.

As how to write test cases, see [README of git.git:t/](https://github.com/git/git/blob/master/t/README).

# Run test

Run all tests (golint, unit test, integration test), please run the following command:

    make test
