#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool hexdump'

. lib/test-lib.sh

test_expect_success 'hexdump (ascii)' '
	printf "hello, world!\n" |
		test-tool hexdump >actual &&
	cat >expect <<-EOF &&
		68 65 6c 6c 6f 2c 20 77 6f 72 6c 64 21 0a 
	EOF
	test_cmp expect actual
'

test_expect_success 'hexdump (unicode)' '
	printf "你好，世界！\n" |
		test-tool hexdump >actual &&
	cat >expect <<-EOF &&
		e4 bd a0 e5 a5 bd ef bc 8c e4 b8 96 e7 95 8c ef bc 81 0a 
	EOF
	test_cmp expect actual
'

test_expect_success 'hexdump (binary)' '
	printf "\001\002\003\004\005\012" |
		test-tool hexdump >actual &&
	cat >expect <<-EOF &&
		01 02 03 04 05 0a 
	EOF
	test_cmp expect actual
'

test_done
