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
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\", \"version\":2}"
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
			"{\"type\":\"agit\", \"version\":2}" \
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
			"{\"type\":\"agit\", \"version\":2}" \
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

test_expect_success "no host/port in ssh_info: bad review url" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		
		----------------------------------------------------------------------
		[FAILED] main/           my/topic1      
		       (bad review URL: file:///path/to/hello/main.git)
		
		EOF
		test_must_fail git-repo upload \
			--assume-yes \
			--no-cache \
			--no-edit \
			--mock-git-push \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"type\":\"agit\", \"version\":2}" \
			>out 2>&1 &&
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
			-e "s#///.*/hello/main.git#///path/to/hello/main.git#g" \
			<out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "manifests url changed to HTTP protocol" '
	(
		cd work/main &&
		git config remote.aone.url https://example.com:1433/hello/main.git
	)
'

test_expect_success "no host/port in ssh_info: use project's http address" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git -c http.extraHeader=AGIT-FLOW: git-repo/n.n.n.n push aone refs/heads/my/topic1:refs/for/Maint/my/topic1
		
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
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_expect_success "manifests url changed to SSH protocol" '
	(
		cd work/main &&
		git config remote.aone.url git@example.com:hello/main.git &&
		echo hack >hack.txt && git add hack.txt &&
		test_tick &&
		git commit -m "hack"
	)
'

test_expect_success "no host/port in ssh_info: use project's ssh address" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project main/ to remote branch Maint:
		  branch my/topic1 ( 2 commit(s)):
		         <hash>
		         <hash>
		to https://example.com (y/N)? Yes
		NOTE: main> will execute command: git push -o oldoid=<hash> aone refs/heads/my/topic1:refs/for/Maint/my/topic1
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
		sed -e "s/[0-9a-f]\{40\}/<hash>/g" -e "s/git-repo\/[^ \"\\]*/git-repo\/n.n.n.n/g" <out >actual &&
		test_cmp expect actual
	)
'

test_done
