#!/bin/sh

test_description="test 'git-repo download' basic"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -u $manifest_url &&
		git-repo sync \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" &&
		git-repo start --all jx/topic
	)
'

test_expect_success "download and checkout" '
	(
		cd work &&
		git-repo download \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			main 12345
	) &&
	(
		cd work/main &&
		echo "Branch: $(git_current_branch)" &&
		git log --pretty="    %s" -2 &&
		git show-ref | cut -c 42- | grep merge-requests
	) >out 2>&1 &&
	sed -e "s/(no branch)/Detached HEAD/g" out >actual &&
	cat >expect<<-EOF &&
	Branch: Detached HEAD
	    New topic
	    Version 0.1.0
	refs/merge-requests/12345/head
	EOF
	test_cmp expect actual
'

test_expect_success "set url to http protocol" '
	(
		cd work/main &&
		git remote set-url aone http://example.com/repository/main.git
	)
'

cat >expect<<-EOF
NOTE: no editor, input data unchanged
##############################################################################
# Step 1: Input your options for code review
#
# Note: Input your options below the comments and keep the comments unchanged,
#       and options which work only for new created code review are hidden.
##############################################################################

# [Issue]       : multiple lines of issue IDs for cross references

# [Reviewer]    : multiple lines of user names as the reviewers for code review

# [Cc]          : multiple lines of user names as the watchers for code review

# [Draft]       : a boolean (yes/no, or true/false) to turn on/off draft mode

# [Private]     : a boolean (yes/no, or true/false) to turn on/off private mode


##############################################################################
# Step 2: Select project and branches for upload
#
# Note: Uncomment the branches to upload, and not touch the project lines
##############################################################################

#
# project ./:
   branch review ( 5 commit(s)) to update code review #12345:
#         <hash>
#         <hash>
#         <hash>
#         <hash>
#         <hash>

NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/repository/main.git refs/heads/review:refs/for-review/12345
NOTE: with extra environment: AGIT_FLOW=1
NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
NOTE: will update-ref refs/merge-requests/12345/head on refs/heads/review, reason: update code review #12345 of http://example.com

----------------------------------------------------------------------
EOF

test_expect_success "git repo upload --single --change 12345" '
	(
		cd work/main &&
		git checkout -b review &&
		git reset --hard aone/master &&
		echo hack >code-review.txt &&
		git add code-review.txt &&
		test_tick &&
		git commit -m "code review from committer"
	) &&
	(
		cd work/main &&
		git-repo upload \
			--single \
			--change 12345 \
			--dryrun \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	) >out 2>&1 &&
	sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
	test_cmp expect actual
'

test_done
