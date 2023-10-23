#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool chmtime'

. lib/test-lib.sh

test_expect_success 'setup' '
	touch -t 200504072213.13 foo &&
	touch -t 200504072213.14 bar &&
	touch -t 200504072213.15 baz
'

test_expect_success 'fail to get timestamp from missing file' '
	test_must_fail test-tool chmtime --get missing-file \
		>actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: failed to stat missing-file, skipping
	EOF
	test_cmp expect actual
'

test_expect_success 'missing timespec to change a file' '
	test_must_fail test-tool chmtime foo >out 2>&1 &&
	grep "^ERROR" out >actual &&
	cat >expect <<-EOF &&
		ERROR: not a base-10 integer: foo
	EOF
	test_cmp expect actual
'

test_expect_success 'missing files' '
	test_must_fail test-tool chmtime >out 2>&1 &&
	grep "error:" out >actual &&
	cat >expect <<-EOF &&
		test-tool chmtime: error: the following arguments are required: files
	EOF
	test_cmp expect actual
'

test_expect_success 'get file timestamp (-g)' '
	test-tool chmtime -g foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112911993
		1112911994
		1112911995
	EOF
	test_cmp expect actual
'

test_expect_success 'get file timestamp (--get)' '
	test-tool chmtime --get foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112911993
		1112911994
		1112911995
	EOF
	test_cmp expect actual
'

test_expect_success 'get file timestamp (-v)' '
	test-tool chmtime -v foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112911993	foo
		1112911994	bar
		1112911995	baz
	EOF
	test_cmp expect actual
'

test_expect_success 'get file timestamp (--verbose)' '
	test-tool chmtime --verbose foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112911993	foo
		1112911994	bar
		1112911995	baz
	EOF
	test_cmp expect actual
'

test_expect_success 'change file timespec (relative to file)' '
	test-tool chmtime -v -100  foo >actual &&
	cat >expect <<-EOF &&
		1112911893	foo
	EOF
	test_cmp expect actual &&

	test-tool chmtime -g +10 bar >actual &&
	cat >expect <<-EOF &&
		1112911893	foo
	EOF

	test-tool chmtime -v foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112911893	foo
		1112912004	bar
		1112911995	baz
	EOF
	test_cmp expect actual
'

test_expect_success 'change file timespec (relative time)' '
	test-tool chmtime =+1000 foo &&
	test-tool chmtime =-1000 bar &&
	test foo -nt bar &&
	test bar -nt baz
'

test_expect_success 'change file timespec (specific time)' '
	test-tool chmtime =1112345678 foo &&
	test-tool chmtime =1123456789 bar &&
	test-tool chmtime --get foo bar baz >actual &&
	cat >expect <<-EOF &&
		1112345678
		1123456789
		1112911995
	EOF
	test_cmp expect actual
'

test_done
