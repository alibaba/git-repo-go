#!/bin/sh

test_description="upload for gerrit remote test"

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
		git-repo sync  \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"gerrit\"}" &&
		git repo start --all my/topic1
	)
'

test_expect_success "check installed hooks" '
	(
		cd work &&
		cat >expect<<-EOF &&
		main
		#!/bin/sh
		project1
		#!/bin/sh
		project1/module1
		#!/bin/sh
		project2
		#!/bin/sh
		EOF
		cat >actual<<-EOF &&
		main
		$(head -1 .repo/projects/main.git/hooks/commit-msg)
		project1
		$(head -1 .repo/projects/projects/app1.git/hooks/commit-msg)
		project1/module1
		$(head -1 .repo/projects/projects/app1/module1.git/hooks/commit-msg)
		project2
		$(head -1 .repo/projects/projects/app2.git/hooks/commit-msg)
		EOF
		test_cmp expect actual &&
		test -L .repo/projects/main.git/hooks/commit-msg &&
		test -L .repo/projects/projects/app1.git/hooks/commit-msg &&
		test -L .repo/projects/projects/app2.git/hooks/commit-msg &&
		test -L .repo/projects/projects/app1/module1.git/hooks/commit-msg
	)
'

test_expect_success "new commit: ready for upload" '
	(
		cd work/main &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file" &&
		cd .. &&
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
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun --drafts (with cache)" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint (draft):
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@ssh.example.com/main.git refs/heads/my/topic1:refs/drafts/Maint
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--draft \
			--dryrun \
			--mock-git-push \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun --drafts (no cache)" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint (draft):
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@ssh.example.com:29418/main.git refs/heads/my/topic1:refs/drafts/Maint
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--draft \
			--dryrun \
			--no-cache \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response "ssh.example.com 29418" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun with reviewers" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push --receive-pack=gerrit receive-pack ssh://committer@ssh.example.com:29418/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1%r=user1,r=user2,r=user3,r=user4,cc=user5,cc=user6,cc=user7,private,wip
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-edit \
			--dryrun \
			--mock-git-push \
			--reviewers user1,user2 \
			--re user3,user4 \
			--cc user5,user6 \
			--cc user7 \
			--wip \
			--private \
			--auto-topic \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" out >actual &&
		test_cmp expect actual
	)
'

test_done
