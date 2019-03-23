#!/bin/sh

test_description="git-repo test log"

. ./lib/sharness.sh

cat >expect <<EOF
WARNING: warn message...
ERROR: error message...
NOTE: note message...
NOTE: hello, world.
EOF

test_expect_success "git-repo test log" '
	git-repo test log >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
WARNING: warn message...
ERROR: error message...
EOF

test_expect_success "git-repo test log -q" '
	git-repo test log -q >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
INFO: info message...
WARNING: warn message...
ERROR: error message...
NOTE: note message...
NOTE: hello, world.
EOF

test_expect_success "git-repo test log -v" '
	git-repo test log -v >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
DEBUG: debug message, with fields...                (my-key=my-value)
DEBUG: debugf message, with fields...               (my-key=my-value)
DEBUG: debugln message, with fields...              (my-key=my-value)
DEBUG: debug message...
INFO: info message...
WARNING: warn message...
ERROR: error message...
NOTE: note message...
NOTE: hello, world.
EOF

test_expect_success "git-repo test log -vv" '
	git-repo test log -vv >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
TRACE: trace message, with fields...                (my-key=my-value)
TRACE: tracef message, with fields...               (key1=value1 key2=value2)
TRACE: traceln message, with fields...              (my-key=my-value)
DEBUG: debug message, with fields...                (my-key=my-value)
DEBUG: debugf message, with fields...               (my-key=my-value)
DEBUG: debugln message, with fields...              (my-key=my-value)
DEBUG: debug message...
INFO: info message...
WARNING: warn message...
ERROR: error message...
NOTE: note message...
NOTE: hello, world.
EOF

test_expect_success "git-repo test log -vvv" '
	git-repo test log -vvv >actual 2>&1 &&
	test_cmp expect actual
'

test_done
