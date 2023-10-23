#!/bin/sh
#
# Copyright (c) 2023 Jiang Xin
#

test_description='Test on test-tool pkt-line'

. lib/test-lib.sh

test_expect_success 'pkt-line from stdin (special packets)' '
	cat <<-EOF |
		0000
		0001
		0002
	EOF
	test-tool pkt-line pack >actual &&
	printf "0000"  >expect &&
	printf "0001" >>expect &&
	printf "0002" >>expect &&
	test_cmp expect actual
'

test_expect_success 'pkt-line from stdin (ascii)' '
	cat <<-EOF |
		Hello, world!
	EOF
	test-tool pkt-line pack >actual &&
	cat >expect <<-EOF &&
		0012Hello, world!
	EOF
	test_cmp expect actual
'

test_expect_success 'pkt-line from stdin (utf-8)' '
	cat <<-EOF |
		Hello, world!
		你好，世界！
	EOF
	test-tool pkt-line pack >actual &&
	cat >expect <<-EOF &&
		0012Hello, world!
		0017你好，世界！
	EOF
	test_cmp expect actual
'

test_expect_success 'pkt-line from args (special packets)' '
	test-tool pkt-line pack 0000 0001 0002 >actual &&
	printf "\n" >>actual &&
	cat >expect <<-EOF &&
		000000010002
	EOF
	test_cmp expect actual
'

test_expect_success 'pkt-line from args (ascii)' '
	test-tool pkt-line pack "Hello, " "world!" >actual &&
	printf "\n" >>actual &&
	cat >expect <<-EOF &&
		000bHello, 000aworld!
	EOF
	test_cmp expect actual
'

test_expect_success 'pkt-line from args (utf-8)' '
	test-tool pkt-line pack "你好，" "世界！" >actual &&
	printf "\n" >>actual &&
	cat >expect <<-EOF &&
		000d你好，000d世界！
	EOF
	test_cmp expect actual
'

test_expect_success 'pack-raw-stdin' '
	cat <<-EOF |
		0001
		Hello, world!
		你好，世界！
		0000
	EOF
	test-tool pkt-line pack-raw-stdin >actual &&
	cat >expect <<-EOF &&
		002f0001
		Hello, world!
		你好，世界！
		0000
	EOF
	test_cmp expect actual
'

test_expect_success 'unpack' '
	cat >input <<-EOF &&
		Hello, world!
		你好，世界！
	EOF
	cat input | test-tool pkt-line pack-raw-stdin >pack.data &&
	cat >expect <<-EOF &&
		0025Hello, world!
		你好，世界！
	EOF
	test_cmp expect pack.data &&
	cat pack.data | test-tool pkt-line unpack >unpack.data &&
	test_cmp input unpack.data
'

test_expect_success 'unpack (has flush packet)' '
	cat >input <<-EOF &&
		Hello, world!
		你好，世界！
	EOF
	printf "0001" >pack.data &&
	cat input | test-tool pkt-line pack-raw-stdin >>pack.data &&
	printf "0000" >>pack.data &&
	cat pack.data | test-tool pkt-line unpack >unpack.data &&
	cat >expect <<-EOF &&
		0001
		Hello, world!
		你好，世界！
		0000
	EOF
	test_cmp expect unpack.data
'

test_expect_failure 'not implement pkt-line command...' '
	test-tool pkt-line unpack-sideband
'

test_expect_success 'bad pkt-line command' '
	test_must_fail test-tool pkt-line bad-command >actual 2>&1 &&
	cat >expect <<-EOF &&
		ERROR: unknown pkt-line cmd: bad-command
	EOF
	test_cmp expect actual
'

test_done
