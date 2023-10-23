#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool date'

. lib/test-lib.sh

test_expect_success 'date is64bit' '
	test_expect_code 1 test-tool date is64bit
'

test_expect_success 'date time_t-is64bit' '
	test_expect_code 1 test-tool date time_t-is64bit
'

test_expect_success 'date getnanos' '
	test-tool date getnanos |
		cut -c1-5 >actual &&
	date +"%s" | cut -c1-5 >expect &&
	test_cmp expect actual
'

test_expect_failure 'not implement date relative...' '
	test-tool date relative
'

test_expect_success 'bad date command' '
	test_must_fail test-tool date bad-command >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: unknown date cmd: bad-command
	EOF
	test_cmp expect actual
'

test_done
