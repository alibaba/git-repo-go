#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool extension module'

. lib/test-lib.sh

test_expect_success 'no TEST_DIRECTORY env' '
	test_must_fail env TEST_DIRECTORY= test-tool missing-cmd >actual 2>&1 &&
	cat >expect <<-EOF &&
		WARN: env TEST_DIRECTORY is not set, unknown command: missing-cmd
	EOF
	test_cmp expect actual
'

test_expect_success 'no extension for missing-cmd' '
	test_path_is_missing "$TEST_DIRECTORY/test-tools" &&
	test_must_fail test-tool missing-cmd >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: unknown command: missing-cmd
	EOF
	test_cmp expect actual
'

test_expect_success 'have extension for missing-cmd' '
	test_when_finished "rm -rf \"$TEST_DIRECTORY/test-tools\"" &&
	mkdir "$TEST_DIRECTORY"/test-tools &&
	(
		cd "$TEST_DIRECTORY"/test-tools &&
		touch __init__.py &&
		cat >missing-cmd.py <<-EOF
			#!/usr/bin/env python

			def Run():
			    print("missing cmd: done")
		EOF
	) &&
	test-tool missing-cmd >actual 2>&1 &&
	cat >expect <<-EOF &&
		missing cmd: done
	EOF
	test_cmp expect actual
'

test_done
