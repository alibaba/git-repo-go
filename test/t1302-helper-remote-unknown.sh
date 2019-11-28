#!/bin/sh

test_description="git-repo helper remote --type agit"

. ./lib/sharness.sh

export PATH="$HOME/bin":$PATH

test_expect_success "setup" '
	(
		mkdir bin && cd bin &&
		cat >git-repo-helper-remote-unknown1 <<-EOF &&
		#!/bin/sh

		git-repo helper remote --type agit "\$@"
		EOF
		chmod a+x git-repo-helper-remote-unknown1 &&
		cat >git-repo-helper-remote-unknown2 <<-EOF &&
		#!/bin/sh

		git-repo helper remote --type gerrit "\$@"
		EOF
		chmod a+x git-repo-helper-remote-unknown2
	)
'

cat >expect <<EOF
{
  "cmd": "git",
  "args": [
    "push",
    "--receive-pack=agit-receive-pack",
    "-o",
    "title=title of code review",
    "-o",
    "description=description of code review",
    "-o",
    "issue=123",
    "-o",
    "reviewers=u1,u2",
    "-o",
    "cc=u3,u4",
    "ssh://git@example.com/test/repo.git",
    "refs/heads/my/topic:refs/for/master/my/topic"
  ],
  "env": [
    "AGIT_FLOW=1"
  ]
}
EOF

test_expect_success "upload command (SSH protocol)" '
	cat <<-EOF |
	{
	  "Description": "description of code review",
	  "DestBranch": "master",
	  "Draft": false,
	  "Issue": "123",
	  "LocalBranch": "my/topic",
	  "People":[
	  	["u1", "u2"],
		["u3", "u4"]
	  ],
	  "ProjectName": "test/repo",
	  "ReviewURL": "ssh://git@example.com",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
  	}	
	EOF
	git-repo helper remote --type unknown1 --upload >out 2>&1 &&
	cat out | jq . >actual &&
	test_cmp expect actual
'

cat >expect <<EOF
refs/changes/45/12345/1
EOF

test_expect_success "download ref" '
	printf "12345\n" | \
	git-repo helper remote --type unknown2 --download >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
Error: cannot find helper 'git-repo-helper-remote-unknown3'
EOF

test_expect_success "cannot find helper program" '
	printf "12345\n" | \
	test_must_fail git-repo helper remote --type unknown3 --download >actual 2>&1 &&
	test_cmp expect actual
'

test_done
