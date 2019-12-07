#!/bin/sh

test_description="upload test"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	# checkout main.git and make it detached
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git clone $main_repo_url main
	)
'

test_expect_success "install git review aliases command" '
	git-repo version &&
	git config alias.peer-review >actual &&
	git config alias.pr >>actual &&
	git config alias.review >>actual &&
	cat >expect <<-EOF &&
	repo upload --single
	repo upload --single
	repo upload --single
	EOF
	test_cmp expect actual
'


test_expect_success "upload error: unsupport url protocol" '
	(
		cd work/main &&
		# git 1.7.10: "git checkout -q" is not really quiet.
		git checkout -q -b jx/topic origin/Maint >/dev/null &&
		test_must_fail git peer-review \
			--no-cache \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	Error: no remote defined for project main
	EOF
	test_cmp expect actual
'

test_expect_success "update remote URL using http protocol" '
	(
		cd work/main &&
		git config remote.origin.url https://example.com/jiangxin/main.git
	)
'

test_expect_success "No commit ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
					"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "New commit in main project" '
	(
		cd work/main &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file"
	)
'

test_expect_success "push.default is unset" '
	(
		cd work/main &&
		test_must_fail git config push.default
	) >actual &&
	cat >expect<<-EOF &&
	EOF
	test_cmp expect actual
'

test_expect_success "will upload one commit for review (http/dryrun/draft/no-edit)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch Maint (draft):
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/drafts/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on https://example.com

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
				--draft \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "push.default has been set to nothing" '
	(
		cd work/main &&
		git config push.default
	) >actual &&
	cat >expect<<-EOF &&
	nothing
	EOF
	test_cmp expect actual
'

test_expect_success "will upload one commit for review (http/dryrun/draft/with edit options)" '
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

		yes

		# [Private]     : a boolean (yes/no, or true/false) to turn on/off private mode


		##############################################################################
		# Step 2: Select project and branches for upload
		#
		# Note: Uncomment the branches to upload, and not touch the project lines
		##############################################################################

		#
		# project ./:
		   branch jx/topic ( 1 commit(s)) to remote branch Maint:
		#         <hash>

		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/drafts/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on https://example.com

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--dryrun \
				--draft \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "will upload one commit for review (http/dryrun)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		EOF
		if git-repo test version --git lt 2.10.0; then
			cat >>expect<<-EOF
			NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic%r=user1,r=user2,r=user3,r=user4,cc=user5,cc=user6,cc=user7,notify=NONE,private,wip
			EOF
		else
			cat >>expect<<-EOF
			NOTE: will execute command: git push --receive-pack=agit-receive-pack -o title=review example -o description={base64}6K+m57uG6K+05piOXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
			EOF
		fi &&
		cat >>expect<<-EOF &&
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on https://example.com

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
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
				--no-emails
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "will upload one commit for review (http/mock-git-push/not-dryrun)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:10022/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":10022, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "published ref will be created" '
	(
		cd work &&
		( cd main && git rev-parse refs/heads/jx/topic ) >expect &&
		( cd main && git rev-parse refs/published/jx/topic ) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload again, no branch ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no change in project . (branch jx/topic) since last upload
		NOTE: no branches ready for upload
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "amend current commit" '
	(
		cd work/main &&
		git commit --amend -m "amend current commit"
	)
'

test_expect_success "update remote URL using ssh port 10022" '
	(
		cd work/main &&
		git config remote.origin.url ssh://git@example.com:10022/jiangxin/main.git
	)
'

test_expect_success "upload to a ssh review url" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -p 10022 git@example.com ssh_info
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:10022 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com:10022/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on ssh://git@example.com:10022

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":10022, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "update remote URL using ssh port 29418" '
	(
		cd work/main &&
		git config remote.origin.url ssh://git@example.com:29418/jiangxin/main.git
	)
'

test_expect_success "no gerrit hooks before review on gerrit" '
	test ! -e work/main/.git/hooks/commit-msg
'

test_expect_success "upload to gerrit ssh review url (assume-no, dryrun, use ssh_info cache)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -p 29418 git@example.com ssh_info
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com:10022/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on ssh://git@example.com:29418

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--assume-yes \
				--no-edit \
				--dryrun \
				--no-cache \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":10022, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "gerrit hooks not installed" '
	test ! -e work/main/.git/hooks/commit-msg
'

test_expect_success "upload to gerrit ssh review url (assume-no, dryrun, no-cache, 29418 is gerrit)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -p 29418 git@example.com ssh_info
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@example.com:29418/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on ssh://git@example.com:29418

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 500 \
				--mock-ssh-info-response ""
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "gerrit hooks installed" '
	test -e work/main/.git/hooks/commit-msg
'

test_expect_success "upload to gerrit ssh review url" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -p 29418 git@example.com ssh_info
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com:10022/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on ssh://git@example.com:29418

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":10022, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "update remote URL with a rcp style URL" '
	(
		cd work/main &&
		git config remote.origin.url git@example.com:jiangxin/main.git
	)
'

test_expect_success "upload to a ssh review using rcp style URL" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh git@example.com ssh_info
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: will update-ref refs/published/jx/topic on refs/heads/jx/topic, reason: review from jx/topic to Maint on ssh://git@example.com

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "create more commits" '
	(
		cd work/main &&
		for i in $(seq 1 10)
		do
			test_tick &&
			git commit --allow-empty -m "commit #$i"
		done
	)
'

test_expect_success "update remote URL back using http protocol" '
	(
		cd work/main &&
		git config remote.origin.url https://example.com/jiangxin/main.git
	)
'

test_expect_success "ATTENTION confirm if there are too many commits for review" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch Maint:
		  branch jx/topic (11 commit(s)):
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		         <hash>
		to https://example.com (y/N)? Yes
		ATTENTION: You are uploading an unusually high number of commits.
		YOU PROBABLY DO NOT MEAN TO DO THIS. (Did you rebase across branches?)
		If you are sure you intend to do this, type '"'"'yes'"'"': Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/jx/topic:refs/for/Maint/jx/topic
		NOTE: with extra environment: AGIT_FLOW=1
		NOTE: with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--no-edit \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_done
