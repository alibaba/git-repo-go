#!/bin/sh

test_description="git-repo helper proto --type agit"

. ./lib/sharness.sh

export PATH="$HOME/bin":$PATH

test_expect_success "setup" '
	(
		mkdir bin && cd bin &&
		cat >git-repo-helper-proto-unknown1 <<-EOF &&
		#!/bin/sh

		git-repo helper proto --type agit "\$@"
		EOF
		chmod a+x git-repo-helper-proto-unknown1 &&
		cat >git-repo-helper-proto-unknown2 <<-EOF &&
		#!/bin/sh

		git-repo helper proto --type gerrit "\$@"
		EOF
		chmod a+x git-repo-helper-proto-unknown2
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
	]
}
EOF

test_expect_success "upload command (SSH protocol, version 0)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "", "Ref": ""},
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
	  "RemoteName": "",
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type unknown1 --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
{
	"cmd": "git",
	"args": [
		"push",
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

test_expect_success "upload command (SSH protocol, version 2)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "", "Ref": ""},
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
	  "RemoteName": "",
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type unknown1 --version 2 --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
refs/changes/45/12345/1
EOF

test_expect_success "download ref" '
	printf "12345\n" | \
	git-repo helper proto --type unknown2 --download >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
Error: cannot find helper 'git-repo-helper-proto-unknown3'
EOF

test_expect_success "cannot find helper program" '
	printf "12345\n" | \
	test_must_fail git-repo helper proto --type unknown3 --download >actual 2>&1 &&
	test_cmp expect actual
'

test_done
