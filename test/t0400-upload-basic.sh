#!/bin/sh

test_description="upload test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work
'

test_expect_success "git-repo init & sync" '
	(
		cd work &&
		git-repo init -u $manifest_url -g all -b Maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "--remote only work with --single" '
	(
		cd work &&
		test_must_fail git-repo upload --remote origin \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
		cat >expect<<-EOF &&
		Error: --remote can be only used with --single
		EOF
		test_cmp expect actual
	)
'

test_expect_success "detached: no branch ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		git-repo upload --mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "new branch: no branch ready for upload" '
	(
		cd work &&
		git repo start --all my/topic1 &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		git-repo upload --mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "new commit" '
	(
		cd work/main &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file"
	)
'

test_expect_success "with new commit, ready for upload (--no-edit)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		ERROR: upload aborted by user
		Error: nothing confirmed for upload
		EOF
		test_must_fail git-repo upload \
			--assume-no \
			--no-edit \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "with new commit, ready for upload (edit push options)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no editor, input data unchanged
		##############################################################################
		# Step 1: Input your options for code review
		#
		# Note: Input your options below the comments and keep the comments unchanged
		##############################################################################
		
		# [Title]       : one line message below as the title of code review
		
		# [Description] : multiple lines of text as the description of code review
		
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
		# project main/:
		   branch my/topic1 ( 1 commit(s)) to remote branch Maint:
		#         <hash>
		
		NOTE: main> will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--dryrun \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "agit-flow proto v2: no agit-receive-pack, and push with environments" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--dryrun \
			--no-cache \
			--no-edit \
			--assume-yes \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "new branch, and do nothing for for upload --cbr" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		git-repo upload --cbr \
			--assume-no \
			--no-cache \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload branch without --cbr" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		test_must_fail git-repo upload \
			--assume-no \
			--no-edit \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		ERROR: upload aborted by user
		Error: nothing confirmed for upload
		EOF
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun --drafts" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint (draft):
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/drafts/Maint/my/topic1
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--draft \
			--dryrun \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		EOF
		if git-repo test version --git lt 2.10.0; then
			cat >>expect<<-EOF
			NOTE: main> will execute command: git push ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1%r=user1,r=user2,r=user3,r=user4,cc=user5,cc=user6,cc=user7,notify=NONE,private,wip
			EOF
		else
			cat >>expect<<-EOF
			NOTE: main> will execute command: git push -o title=review example -o description={base64}6K+m57uG6K+05piOXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
			EOF
		fi &&
		cat >>expect<<-EOF &&
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--dryrun \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			--reviewers user1,user2 \
			--re user3,user4 \
			--cc user5,user6 \
			--cc user7 \
			--title "review example" \
			--description "详细说明\n...\n" \
			--private \
			--wip \
			--no-emails \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "mock-git-push, but do update-ref for upload" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push ssh://git@ssh.example.com/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "check update-ref" '
	(
		cd work &&
		( cd main && git rev-parse refs/heads/my/topic1 ) >expect &&
		( cd main && git rev-parse refs/published/my/topic1/Maint ) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload again, no branch ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no change in project main (branch my/topic1) since last upload
		NOTE: no branches ready for upload
		EOF
		git-repo upload \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
		test_cmp expect actual
	)
'

test_done
