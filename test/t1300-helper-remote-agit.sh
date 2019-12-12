#!/bin/sh

test_description="git-repo helper proto --type agit"

. ./lib/sharness.sh

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
		"origin",
		"refs/heads/my/topic:refs/for/master/my/topic"
	]
}
EOF

test_expect_success "upload command (SSH protocol, verison 0)" '
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
	  "RemoteName": "origin",
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --upload >actual 2>&1 &&
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
		"origin",
		"refs/heads/my/topic:refs/for/master/my/topic"
	],
	"env": [
		"AGIT_FLOW=1"
	]
}
EOF

test_expect_success "upload command (SSH protocol, verison 2)" '
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
	  "RemoteName": "origin",
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --version 2 --upload >actual 2>&1 &&
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
		"origin",
		"refs/heads/my/topic:refs/drafts/master/my/topic"
	],
	"env": [
		"AGIT_FLOW=1"
	]
}
EOF

test_expect_success "upload command (SSH protocol, draft, version 2)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "", "Ref": ""},
	  "Description": "description of code review",
	  "DestBranch": "master",
	  "Draft": true,
	  "Issue": "123",
	  "LocalBranch": "my/topic",
	  "People":[
		["u1", "u2"],
		["u3", "u4"]
	  ],
	  "ProjectName": "test/repo",
	  "RemoteName": "origin",
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --version 2 --upload >actual 2>&1 &&
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
		"example",
		"refs/heads/my/topic:refs/for/master/my/topic"
	],
	"gitconfig": [
		"http.extraHeader=\"AGIT-FLOW: 1\""
	]
}
EOF

test_expect_success "upload command (HTTP protocol, version 0)" '
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
	  "RemoteName": "example",
	  "RemoteURL": "https://example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --upload >actual 2>&1 &&
	test_cmp expect actual
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
		"ssh://git@example.com:29418/test/repo.git",
		"refs/heads/my/topic:refs/for-review/12345"
	]
}
EOF

test_expect_success "upload command (SSH protocol with code review ID, version 0)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "12345", "Ref": "refs/merge-requests/12345"},
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
	  "RemoteURL": "ssh://git@example.com:29418/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --upload >actual 2>&1 &&
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
		"ssh://git@example.com:29418/test/repo.git",
		"refs/heads/my/topic:refs/for-review/12345"
	],
	"env": [
		"AGIT_FLOW=1"
	]
}
EOF

test_expect_success "upload command (SSH protocol with code review ID, version 2)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "12345", "Ref": "refs/merge-requests/12345"},
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
	  "RemoteURL": "ssh://git@example.com:29418/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --version 2 --upload >actual 2>&1 &&
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
		"origin",
		"refs/heads/my/topic:refs/for-review/12345"
	],
	"gitconfig": [
		"http.extraHeader=\"AGIT-FLOW: 1\""
	]
}
EOF

test_expect_success "upload command (HTTP protocol with code review ID, draft)" '
	cat <<-EOF |
	{
	  "CodeReview": {"ID": "12345", "Ref": "refs/merge-requests/12345"},
	  "Description": "description of code review",
	  "DestBranch": "master",
	  "Draft": true,
	  "Issue": "123",
	  "LocalBranch": "my/topic",
	  "People":[
		["u1", "u2"],
		["u3", "u4"]
	  ],
	  "ProjectName": "test/repo",
	  "RemoteName": "origin",
	  "RemoteURL": "http://example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type agit --version 2 --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
refs/merge-requests/12345/head
EOF

test_expect_success "download ref" '
	printf "12345\n" | \
	git-repo helper proto --type agit --download >actual 2>&1 &&
	test_cmp expect actual
'

test_done
