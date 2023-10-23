#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool env-helper'

. lib/test-lib.sh

test_expect_success 'true: env set as true' '
	env TEST_ENV_HELPER_VAR1=true \
		test-tool env-helper --type bool TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'true: env set as yes' '
	env TEST_ENV_HELPER_VAR1=yes \
		test-tool env-helper --type bool TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'true: env default to on' '
	test-tool env-helper --type bool --default on \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'true: env default to 1' '
	test-tool env-helper --type bool --default 1 \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'true: env default to 255' '
	test-tool env-helper --type bool --default 255 \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'true: env default to -1' '
	test-tool env-helper --type bool --default -1 \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		true
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env not set' '
	test_expect_code 1 \
		test-tool env-helper --type bool \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'false: empty env is false' '
	test_expect_code 1 \
		env TEST_ENV_HELPER_VAR1= \
		test-tool env-helper --type bool TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env set to false' '
	test_expect_code 1 \
		env TEST_ENV_HELPER_VAR1=false \
		test-tool env-helper --type bool TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env set to no' '
	test_expect_code 1 \
		env TEST_ENV_HELPER_VAR1=no \
		test-tool env-helper --type bool TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env default to off' '
	test_expect_code 1 \
		test-tool env-helper --type bool --default off \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env default to 0' '
	test_expect_code 1 \
		test-tool env-helper --type bool --default 0 \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		false
	EOF
	test_cmp expect actual
'

test_expect_success 'error: env with invalid default' '
	test_expect_code 128 \
		test-tool env-helper --type bool --default invalid \
		TEST_ENV_HELPER_VAR1 >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: bad boolean value ${SQ}invalid${SQ}
	EOF
	test_cmp expect actual
'

test_expect_success 'false: env is invalid' '
	test_expect_code 128 \
		env TEST_ENV_HELPER_VAR1=invalid \
		test-tool env-helper --type bool \
		TEST_ENV_HELPER_VAR1 >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: bad boolean value ${SQ}invalid${SQ}
	EOF
	test_cmp expect actual
'

test_expect_success 'ulong env not set as 0' '
	test_expect_code 1 \
		test-tool env-helper --type ulong \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		0
	EOF
	test_cmp expect actual
'

test_expect_success 'ulong env: 255' '
	env TEST_ENV_HELPER_VAR1=255 \
		test-tool env-helper --type ulong \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		255
	EOF
	test_cmp expect actual
'

test_expect_success 'ulong env: 1k' '
	env TEST_ENV_HELPER_VAR1=1k \
		test-tool env-helper --type ulong \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		1024
	EOF
	test_cmp expect actual
'

test_expect_success 'ulong env: 1m' '
	test-tool env-helper --type ulong --default=1m \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		1048576
	EOF
	test_cmp expect actual
'

test_expect_success 'ulong env: 1g' '
	env TEST_ENV_HELPER_VAR1=1g test-tool env-helper \
		--type ulong --default=1k \
		TEST_ENV_HELPER_VAR1 >actual &&
	cat >expect <<-EOF &&
		1073741824
	EOF
	test_cmp expect actual
'

test_expect_success 'error: ulong is bad-number' '
	test_expect_code 129 \
		env TEST_ENV_HELPER_VAR1=100-bad-number \
		test-tool env-helper --type ulong --default=1k \
		TEST_ENV_HELPER_VAR1 >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: failed to parse ulong number ${SQ}100-bad-number${SQ}
	EOF
	test_cmp expect actual
'

test_done
