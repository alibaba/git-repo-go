#!/bin/sh

test_description="upload with different pushurl test"

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
			"{\"pushurl\":\"https://git.example.com/git\", \"type\":\"agit\", \"version\":2}"
	)
'

test_expect_success "new commit in my/topic1" '
	(
		cd work/main &&
		git repo start --all my/topic1 &&
		echo hack >topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: new file"
	)
'

test_expect_success "use pushurl in manifest remotes" '
	(
		cd work &&
		mkdir .repo/local_manifests &&
		cat >.repo/local_manifests/test.xml <<-EOF &&
		<?xml version="1.0" encoding="UTF-8"?>
		<manifest>
		  <remote  name="aone"
		   alias="origin"
		   fetch="."
		   pushurl="ssh://{email}@aone.example.com/agit"
		   override="true"
		   review="https://example.com" />
		  <remote  name="driver"
		   fetch=".."
		   pushurl="ssh://{email}@driver.example.com/agit"
		   override="true"
		   review="https://example.com" />
		</manifest>
		EOF
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push ssh://committer@aone.example.com/agit/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--assume-yes \
			--no-cache \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
			-e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" \
			-e "s#///.*/hello/main.git#///path/to/hello/main.git#g" \
			<out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "new commit in my/topic1" '
	(
		cd work/main &&
		echo hack2 >>topic1.txt &&
		git add topic1.txt &&
		test_tick &&
		git commit -m "topic1: hack2"
	)
'

test_expect_success "pushurl in manifest override ssh-info response" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 2 commit(s)):
		         <hash>
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push -o oldoid=<hash> ssh://committer@aone.example.com/agit/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> with extra environment: AGIT_FLOW=git-repo/n.n.n.n
		NOTE: main> with extra environment: GIT_SSH_COMMAND=ssh -o SendEnv=AGIT_FLOW
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--dryrun \
			--assume-yes \
			--no-cache \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"type\":\"agit\", \"version\":2, \"pushurl\":\"https://<email>@git.example.com/agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
			-e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" \
			-e "s#///.*/hello/main.git#///path/to/hello/main.git#g" \
			<out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "remove pushurl in manifest" '
	rm work/.repo/local_manifests/test.xml
'

test_expect_success "use pushurl in ssh-info response" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 2 commit(s)):
		         <hash>
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git -c http.extraHeader=AGIT-FLOW: git-repo/n.n.n.n push -o oldoid=<hash> https://committer@git.example.com/agit/main.git refs/heads/my/topic1:refs/for/Maint/my/topic1
		NOTE: main> will update-ref refs/published/my/topic1/Maint on refs/heads/my/topic1, reason: review from my/topic1 to Maint on https://example.com
		
		----------------------------------------------------------------------
		EOF
		git-repo upload \
			--dryrun \
			--assume-yes \
			--no-cache \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"type\":\"agit\", \"version\":2, \"pushurl\":\"https://<email>@git.example.com/agit\"}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
			-e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" \
			-e "s#///.*/hello/main.git#///path/to/hello/main.git#g" \
			<out >actual &&
		test_cmp expect actual
	)
'

test_done
