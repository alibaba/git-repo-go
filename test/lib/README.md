Test Lib From Git Core
======================

Test-lib is a test framework developed by Junio and is specifically
designed for the Git project. It allows us to write a test suite
using shell script, which contains a collection of test cases. The
output of each test suite is presented in TAP ([Test Anything
Protocol]) format. We can use test-lib or any other TAP harness
programs (e.g., prove) to run and analyze the output of the test
suites.

In order to reuse the git test framework into other projects,
the [sharness project] made a successful attempt. However, it is
based on an outdated version of git (v1.7.9), which results in
bugs and missing new features. For example:

 * Commit d88785e424 (test-lib: set `BASH_XTRACEFD` automatically,
   2016-05-11) and commit a5bf824f3b4d (t: prevent '-x' tracing
   from interfering with test helpers' stderr, 2018-02-25) of the Git
   project addressed bugs when we run test suites with "-x" option.

 * Commit 0445e6f0a1 (test-lib: '--run' to run only specific tests,
   2014-04-30) provided better control of the set of tests to run.

 * Commit 92b269f5c5 (test-lib: turn on `GIT_TEST_CHAIN_LINT`
   by default, 2015-04-22) turned on chain-lint by default to
   prevent the accidental omission of "&&" between statements in
   test cases. Additionally, Git offers more linter tools such as
   "chainlint.pl" to help write correct test cases.

 * The latest version of the Git project includes numerous test
   helpers that are not present in sharness. These helpers
   provide a more comprehensive and efficient testing developing
   tools. E.g.: `test_bool_env`, `test_cmp_bin`, `test_commit`,
   `test_config`, `test_env`, `test_file_not_empty`, `test_file_size`,
   `test_line_count`, `test_oid`, `test_path_exists`,
   `test_path_is_executable`, `test_path_is_missing`, `test_tick`,
   `write_script`, etc.

In order to reuse the latest test framework of the Git project and
easy to maintain, use the following strategies:

1. Use [git-filter-repo] to export test-lib related files and their
   commit histories from the Git project. The resulting tailored
   commits are saved in the branch named "git-test-lib".

2. The test-lib test framework relies on a helper program named
   "test-tool", which is written in C. To use test-lib without the
   need for C compilation, re-implemented part of the "test-tool"
   subcommands in Python.

3. Make some modifications to test-lib, such as sourcing "test-lib.sh"
   from a subdirectory other than the current directory. Some
   modifications are borrowed from Sharness.

4. The "git-test-lib" branch is continuously updated with the Git
   project, and the master branch will be rebased on it, so the master
   branch is not stable. For instructions on how to tailor the
   test-lib framework from the Git project to update the "git-test-lib"
   branch, please refer to the last section.


Install test-lib
----------------

To create test suites in shell scripts powered by test-lib, you can
follow these steps:

1. Set up a directory (such as "test") to save test suites and
   files of test-lib.

        $ mkdir test

2. Clone or copy the test-lib repository inside the test directory.

        $ cd test
        $ git clone https://github.com/jiangixn/test-lib lib

4. Copy files from the example test directory.

        $ cp lib/test-example/.gitignore .
        $ cp lib/test-example/.gitattributes .
        $ cp lib/test-example/Makefile .

4. Start writing a test suite powered by test-lib, make sure it
   sources the "test-lib.sh" file from test-lib. Refer to the example
   test suite (e.g. "test/t0001-test-tool-chmtime.sh") to write
   your own test suite:

        test_description='My first test suite'

        . lib/test-lib.sh


Usage of test-lib
-----------------

As for how to use test-lib to write, run and manage your test suites,
please see the documentation [README.git] for reference.


Filtering test-lib from the Git project
---------------------------------------

"test-lib" is part of the Git project and is stored in the "t/"
subdirectory. We use "git-filter-repo" to export test-lib related
files to the root directory of this project. We repeat this
periodically to save historical commits of test-lib in the
"git-test-lib" branch of this project. The master branch of this
project will merge with or rebase onto the "git-test-lib" branch.
As how to filter and export test-lib from the Git project, see the
steps below.

1. Make a fresh clone of the Git project before filtering.

        $ git clone --single-branch --no-tags \
          https://github.com/git/git.git \
          git-test-lib

        $ cd git-test-lib

2. Use git-filter-repo to filter and export test-lib.

        $ git filter-repo \
          --preserve-commit-encoding \
          --prune-degenerate always \
          --path COPYING \
          --path shared.mak \
          --path t/.gitattributes \
          --path t/.gitignore \
          --path t/Makefile \
          --path t/README \
          --path t/aggregate-results.sh \
          --path t/chainlint/ \
          --path t/chainlint.pl \
          --path t/check-non-portable-shell.pl \
          --path t/oid-info/ \
          --path t/perf/.gitignore \
          --path t/perf/Makefile \
          --path t/perf/README \
          --path t/perf/aggregate.perl \
          --path t/perf/config \
          --path t/perf/min_time.perl \
          --path t/perf/perf-lib.sh \
          --path t/perf/p0000-perf-lib-sanity.sh \
          --path t/perf/run \
          --path t/test-lib-functions.sh \
          --path t/test-lib-github-workflow-markup.sh \
          --path t/test-lib-junit.sh \
          --path t/test-lib.sh \
          --path t/lib-subtest.sh \
          --path t/t0000-basic.sh \
          --path-rename t/:


[Test Anything Protocol]: http://testanything.org/
[sharness project]: https://github.com/felipec/sharness
[git-filter-repo]: https://github.com/newren/git-filter-repo
[README.git]: ./README.git
