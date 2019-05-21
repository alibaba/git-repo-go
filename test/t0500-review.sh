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

test_expect_success "install git review alias command" '
	git-repo --version &&
	git config alias.review >actual &&
	cat >expect <<-EOF &&
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "git review -h" '
	cat >expect<<-EOF &&
	'"'"'review'"'"' is aliased to '"'"'repo upload --single'"'"'
	EOF
	git review -h >/dev/null 2>actual &&
	test_cmp expect actual
'

test_expect_success "upload error: not in a branch" '
	(
		cd work &&
		cat >expect<<-EOF &&
		FATAL: upload failed: not in a branch
		
		Please run command "git checkout -b <branch>" to create a new branch.
		EOF
		cd main &&
		git checkout HEAD^0 &&
		cd .. &&
		test_must_fail git -C main review >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "upload error: cannot find track branch" '
	(
		cd work &&
		git -C main checkout -b my/topic-test &&
		cat >expect<<-EOF &&
		FATAL: upload failed: cannot find tracking branch
		
		Please run command "git branch -u <upstream>" to track a remote branch. E.g.:
		
		    git branch -u origin/master
		EOF
		test_must_fail git -C main review >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "upload error: no remote URL" '
	(
		cd work &&
		git -C main branch -u origin/master &&
		oldurl=$(git -C main config remote.origin.url) &&
		git -C main config --unset remote.origin.url &&
		cat >expect<<-EOF &&
		FATAL: upload failed: unknown URL for remote: origin
		EOF
		test_must_fail git -C main review >actual 2>&1 &&
		test_cmp expect actual &&
		git -C main config remote.origin.url $oldurl
	)
'

test_expect_success "upload error: unknown URL protocol" '
	(
		cd work &&
		cat >expect<<-EOF &&
		FATAL: cannot find review URL from '"'"'file:///path/of/main.git'"'"'
		EOF
		test_must_fail git -C main review >out 2>&1 &&
		sed -e "s#///.*/main.git#///path/of/main.git#" <out >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "update remote URL using http protocol" '
	(
		cd work &&
		git -C main config remote.origin.url https://example.com/jiangxin/main.git
	)
'

test_expect_success "No commit ready for upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		NOTE: no branches ready for upload
		EOF
		git -C main review \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>actual 2>&1 &&
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
		git -C main review \
			--assume-yes \
			--no-edit \
			--dryrun \
			--draft \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
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
		git -C main review \
			--assume-yes \
			--dryrun \
			--draft \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
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
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o title=review example -o description={base64}6K+m57uG6K+05piOXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
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
			--no-emails \
			>out 2>&1 &&
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
		git -C main review \
			--assume-yes \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":10022, \"type\":\"agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "published ref will be created" '
	(
		cd work &&
		git -C main rev-parse refs/heads/my/topic-test >expect &&
		git -C main rev-parse refs/published/my/topic-test >actual &&
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
		git -C main review \
			--assume-yes \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
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
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:10022 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@example.com:10022/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:10022
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--no-edit \
			--dryrun \
			>out 2>&1 &&
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

test_expect_success "upload to gerrit ssh review url" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com:29418 (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=gerrit receive-pack ssh://git@example.com:29418/jiangxin/main.git refs/heads/my/topic-test:refs/for/master
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com:29418
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--no-edit \
			--dryrun \
			>out 2>&1 &&
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
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to ssh://git@example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on ssh://git@example.com
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--no-edit \
			--dryrun \
			>out 2>&1 &&
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
		cd work &&
		git -C main config remote.origin.url https://example.com/jiangxin/main.git
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
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test

		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" <out >actual &&
		test_cmp expect actual
	)
'

test_done
