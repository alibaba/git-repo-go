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
	git-repo --version &&
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

test_expect_success "upload error: no branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: upload failed: not in a branch

		Please run command "git checkout -b <branch>" to create a new branch.
		EOF
		cd main &&
		git checkout HEAD^0 &&
		cd .. &&
		(
			cd main &&
			test_must_fail git peer-review
		) >out 2>&1 &&
		sed -e "s#file:///.*#file:///path/of/repo.git#" out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload error: unsupport url protocol" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: no remote defined for project main
		EOF
		(
			cd main &&
			git checkout -q master &&
			test_must_fail git peer-review \
				--no-cache
		) >out 2>&1 &&
		sed -e "s#file:///.*#file:///path/of/repo.git#" out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "update remote URL using http protocol" '
	(
		cd work/main &&
		git config remote.origin.url https://example.com/jiangxin/main.git
	)
'

test_expect_success "upload error: not in a branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: upload failed: not in a branch

		Please run command "git checkout -b <branch>" to create a new branch.
		EOF
		(
			cd main &&
			git checkout -q HEAD^0 &&
			test_must_fail git peer-review  \
				--no-cache \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
					"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "upload error: cannot find track branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: upload failed: cannot find tracking branch

		Please run command "git branch -u <upstream>" to track a remote branch. E.g.:

		    git branch -u origin/master
		
		Or give the following options when uploading:
		
		    --dest <dest-branch> [--remote <remote>]
		EOF
		(
			cd main &&
			git checkout -q -b my/topic-test &&
			test_must_fail git peer-review \
				--no-cache \
				--mock-git-push \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
					"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "upload error: no remote URL" '
	(
		cd work &&
		cat >expect<<-EOF &&
		WARNING: no URL defined for remote: origin
		Error: no remote defined for project main
		EOF
		(
			cd main &&
			git checkout -q my/topic-test &&
			git config branch.my/topic-test.remote origin &&
			git config branch.my/topic-test.merge refs/heads/master &&
			git config --unset remote.origin.url
			test_must_fail git peer-review \
				--no-cache
		) >actual 2>&1 &&
		test_cmp expect actual
	)
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

test_expect_success "No commit ready for upload (use cached ssh_info)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		(
			cd main &&
			git peer-review \
				--mock-git-push
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
		Upload project (jiangxin/main) to remote branch master (draft):
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/drafts/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on https://example.com

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
		   branch my/topic-test ( 1 commit(s)) to remote branch master:
		#         <hash>

		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/drafts/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on https://example.com

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
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		EOF
		if git-repo test version --git lt 2.10.0; then
			cat >>expect<<-EOF
			NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test%r=user1,r=user2,r=user3,r=user4,cc=user5,cc=user6,cc=user7,notify=NONE,private,wip
			EOF
		else
			cat >>expect<<-EOF
			NOTE: will execute command: git push --receive-pack=agit-receive-pack -o title=review example -o description={base64}6K+m57uG6K+05piOXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
			EOF
		fi &&
		cat >>expect<<-EOF &&
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on https://example.com

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
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:10022/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test

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
		( cd main && git rev-parse refs/heads/my/topic-test ) >expect &&
		( cd main && git rev-parse refs/published/my/topic-test ) >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload again, no branch ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no change in project . (branch my/topic-test) since last upload
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

test_expect_success "upload to a ssh review url (no ssh_info cache)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -l git -p 10022 example.com ssh_info
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
		
		yes
		
		# [Private]     : a boolean (yes/no, or true/false) to turn on/off private mode
		
		
		##############################################################################
		# Step 2: Select project and branches for upload
		#
		# Note: Uncomment the branches to upload, and not touch the project lines
		##############################################################################
		
		#
		# project ./:
		   branch my/topic-test ( 1 commit(s)) to remote branch master:
		#         <hash>
		
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/drafts/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:10022

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--no-cache \
				--assume-yes \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
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
		NOTE: mock executing: ssh -l git -p 29418 example.com ssh_info
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:29418

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--assume-yes \
				--no-edit \
				--dryrun \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "gerrit hooks not installed" '
	test ! -e work/main/.git/hooks/commit-msg
'

test_expect_success "upload to gerrit ssh review url (assume-no, dryrun, no-cache)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: mock executing: ssh -l git -p 29418 example.com ssh_info
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@ssh.example.com:29418/jiangxin/main.git refs/heads/my/topic-test:refs/for/master
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:29418

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
				"ssh.example.com 29418"
		) >out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "gerrit hooks installed" '
	test -e work/main/.git/hooks/commit-msg
'

test_expect_success "upload to gerrit ssh review url (use ssh_info cache)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@ssh.example.com:29418/jiangxin/main.git refs/heads/my/topic-test:refs/for/master
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:29418

		----------------------------------------------------------------------
		EOF
		(
			cd main &&
			git peer-review \
				--assume-yes \
				--no-edit \
				--dryrun
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
		NOTE: mock executing: ssh -l git example.com ssh_info
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com

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
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
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
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test (11 commit(s)):
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
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o oldoid=<hash> ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test

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
