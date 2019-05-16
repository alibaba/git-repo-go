#!/bin/sh

test_description="upload test"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -u $manifest_url -g all -b maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
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
		git -C main branch -u aone/master &&
		oldurl=$(git -C main config remote.aone.url) &&
		git -C main config --unset remote.aone.url &&
		cat >expect<<-EOF &&
		FATAL: upload failed: unknown URL for remote: aone
		EOF
		test_must_fail git -C main review >actual 2>&1 &&
		test_cmp expect actual &&
		git -C main config remote.aone.url $oldurl
	)
'

test_expect_success "upload error: unknown URL protocol" '
	(
		cd work &&
		cat >expect<<-EOF &&
		FATAL: cannot find review URL from '"'"'file:///home/jiangxin/work/git-repo/git-repo/test/test-repositories/hello/main.git'"'"'
		EOF
		test_must_fail git -C main review >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_expect_success "update remote URL using http protocol" '
	(
		cd work &&
		git -C main config remote.aone.url https://example.com/jiangxin/main.git
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

test_expect_success "will upload one commit for review (dryrun, draft)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master (draft):
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/jiangxin/main.git refs/heads/my/topic-test:refs/drafts/master/my/topic-test
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
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "will upload one commit for review (dryrun)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o title={base64}cmV2aWV3IGV4YW1wbGU= -o description={base64}cmV2aWV3IGRlc2NyaXB0aW9uXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com:22/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		NOTE: will update-ref refs/published/my/topic-test on refs/heads/my/topic-test, reason: review from my/topic-test to master on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--dryrun \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			--reviewers user1,user2 \
			--re user3,user4 \
			--cc user5,user6 \
			--cc user7 \
			--title "review example" \
			--description "review description\n...\n" \
			--private \
			--wip \
			--no-emails \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "will upload one commit for review (mock-git-push, not dryrun)" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project (jiangxin/main) to remote branch master:
		  branch my/topic-test ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test
		
		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
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
		NOTE: no branches ready for upload
		EOF
		git -C main review \
			--assume-yes \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
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
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/jiangxin/main.git refs/heads/my/topic-test:refs/for/master/my/topic-test

		----------------------------------------------------------------------
		EOF
		git -C main review \
			--assume-yes \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_done
