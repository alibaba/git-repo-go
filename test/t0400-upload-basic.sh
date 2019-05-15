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
		git-repo init -u $manifest_url -g all -b maint &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
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

test_expect_success "new commit: ready for upload" '
	(
		cd work/main &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file" &&
		cd .. &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		test_must_fail git-repo upload --assume-no --mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
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
		git-repo upload --cbr --assume-no --mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload branch without --cbr" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		test_must_fail git-repo upload --assume-no --mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun --drafts" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch  (draft):
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/main.git refs/heads/my/topic1:refs/drafts/maint/my/topic1
		NOTE: will update-ref refs/published/my/topic1 on refs/heads/my/topic1, reason: review from my/topic1 to maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload --assume-yes --draft --dryrun \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload --dryrun" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack -o title={base64}cmV2aWV3IGV4YW1wbGU= -o description={base64}cmV2aWV3IGRlc2NyaXB0aW9uXG4uLi5cbg== -o reviewers=user1,user2,user3,user4 -o cc=user5,user6,user7 -o notify=no -o private=yes -o wip=yes ssh://git@ssh.example.com:22/main.git refs/heads/my/topic1:refs/for/maint/my/topic1
		NOTE: will update-ref refs/published/my/topic1 on refs/heads/my/topic1, reason: review from my/topic1 to maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload --assume-yes --dryrun \
			--mock-git-push \
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

test_expect_success "mock-git-push, but do update-ref for upload" '
	(
		cd work &&
		git repo start --all my/topic2 &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/main.git refs/heads/my/topic1:refs/for/maint/my/topic1
		
		----------------------------------------------------------------------
		EOF
		git-repo upload --assume-yes \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 | sed -e "s/[0-9a-f]\{40\}/<hash>/g" >actual &&
		test_cmp expect actual
	)
'

test_expect_success "check update-ref" '
	(
		cd work &&
		git -C main rev-parse refs/heads/my/topic1 >expect &&
		git -C main rev-parse refs/published/my/topic1 >actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload again, no branch ready for upload" '
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

test_done
