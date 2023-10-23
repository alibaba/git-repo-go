#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool path-utils'

. lib/test-lib.sh

test_expect_success 'setup' '
	cat >foo <<-EOF &&
		Hello, world!
		你好，世界！
	EOF
	cat foo >bar &&
	cat foo >>bar
'

test_expect_success 'file-size' '
	test-tool path-utils file-size foo bar >actual &&
	cat >expect <<-EOF &&
		33
		66
	EOF
	test_cmp expect actual
'

test_expect_success 'file-size on missing files' '
	test_must_fail test-tool path-utils file-size \
		missing-file foo bar >actual 2>error &&
	cat >expect <<-EOF &&
		33
		66
	EOF
	test_cmp expect actual &&
	cat >expect <<-EOF &&
		ERROR: cannot stat ${SQ}missing-file${SQ}
	EOF
	test_cmp expect error

'

test_expect_success 'skip-n-bytes (mising file)' '
	test_must_fail test-tool path-utils \
		skip-n-bytes missing-file 100 >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: could not open ${SQ}missing-file${SQ}
	EOF
	test_cmp expect actual
'

test_expect_success 'skip-n-bytes (wrong number of args)' '
	test_must_fail test-tool path-utils \
		skip-n-bytes foo >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: need file and number of bytes to skip
	EOF
	test_cmp expect actual
'

test_expect_success 'skip-n-bytes (wrong order of args)' '
	test_must_fail test-tool path-utils \
		skip-n-bytes 0 foo >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: ${SQ}foo${SQ} is not a number, wrong order of args?
	EOF
	test_cmp expect actual
'

test_expect_success 'skip-n-bytes (0 byte)' '
	test-tool path-utils skip-n-bytes foo 0 >actual &&
	test_cmp foo actual
'

test_expect_success 'skip-n-bytes (100 byte on foo, nothing left)' '
	test-tool path-utils skip-n-bytes foo 100 >actual &&
	cat >expect <<-EOF &&
	EOF
	test_cmp expect actual
'

test_expect_success 'skip-n-bytes (33 byte on bar)' '
	test-tool path-utils skip-n-bytes bar 33 >actual &&
	test_cmp foo actual
'

test_expect_success 'skip-n-bytes (56 byte on bar)' '
	test-tool path-utils skip-n-bytes bar 56 >actual &&
	cat >expect <<-EOF &&
		世界！
	EOF
	test_cmp expect actual
'

test_expect_failure 'not implement path-utils command...' '
	test-tool path-utils real_path
'

test_expect_success 'bad path-utils command' '
	test_must_fail test-tool path-utils bad-command >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: unknown path-utils cmd: bad-command
	EOF
	test_cmp expect actual
'

test_done
