#!/bin/sh

test_description="test git-repo version"

. ./lib/sharness.sh

test_expect_success "git-repo version output test" '
	git-repo version >out &&
	grep "^git-repo version" out |
		sed -e "s/[0-9][0-9]*\.[0-9][0-9]*\.[0-9].*$/N.N.N/" \
		>actual &&
	cat >expect <<-EOF &&
	git-repo version N.N.N
	EOF
	test_cmp expect actual
'

test_done
