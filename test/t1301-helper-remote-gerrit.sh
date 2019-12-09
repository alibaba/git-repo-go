#!/bin/sh

test_description="git-repo helper proto --type gerrit"

. ./lib/sharness.sh

cat >expect <<EOF
{
	"cmd": "git",
	"args": [
		"push",
		"--receive-pack=gerrit receive-pack",
		"origin",
		"refs/heads/my/topic:refs/for/master%r=u1,r=u2,cc=u3,cc=u4"
	]
}
EOF

test_expect_success "upload command (SSH protocol)" '
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
	  "RemoteURL": "ssh://git@example.com:29418/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type gerrit --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
{
	"cmd": "git",
	"args": [
		"push",
		"--receive-pack=gerrit receive-pack",
		"ssh://git@example.com/test/repo.git",
		"refs/heads/my/topic:refs/drafts/master%r=u1,r=u2,cc=u3,cc=u4"
	]
}
EOF

test_expect_success "upload command (SSH protocol, draft)" '
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
	  "RemoteURL": "ssh://git@example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type gerrit --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
{
	"cmd": "git",
	"args": [
		"push",
		"https://example.com/test/repo.git",
		"refs/heads/my/topic:refs/for/master%r=u1,r=u2,cc=u3,cc=u4"
	]
}
EOF

test_expect_success "upload command (HTTP protocol)" '
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
	  "RemoteURL": "https://example.com/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	git-repo helper proto --type gerrit --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
Error: Change code review by ID is not allowed in Gerrit
EOF

test_expect_success "upload command (SSH protocol with code review ID)" '
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
	  "RemoteName": "origin",
	  "RemoteURL": "ssh://git@example.com:29418/test/repo.git",
	  "Title": "title of code review",
	  "UserEmail": "Jiang Xin <worldhello.net@gmail.com>",
	  "Version": 1
	}
	EOF
	test_must_fail git-repo helper proto --type gerrit --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
Error: Change code review by ID is not allowed in Gerrit
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
	test_must_fail git-repo helper proto --type gerrit --upload >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
WARNING: Patch ID should not be 0, set it to 1
refs/changes/45/12345/1
EOF

test_expect_success "download ref (no patch)" '
	printf "12345\n" | \
	git-repo helper proto --type gerrit --download >actual 2>&1 &&
	test_cmp expect actual
'

cat >expect <<EOF
refs/changes/45/12345/2
EOF

test_expect_success "download ref (with patch)" '
	printf "12345 2\n" | \
	git-repo helper proto --type gerrit --download >actual 2>&1 &&
	test_cmp expect actual
'

test_done
