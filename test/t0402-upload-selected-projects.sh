#!/bin/sh

test_description="upload with args to select projects"

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
		git-repo sync
	)
'

test_expect_success "create commits" '
	(
		cd work &&
		git repo start --all my/topic1 &&
		test_tick &&
		(
			cd main &&
			echo hack >topic1.txt &&
			git add topic1.txt &&
			git commit -m "topic1: new file"
		) &&
		test_tick &&
		(
			cd projects/app1 &&
			echo hack >topic1.txt &&
			git add topic1.txt &&
			test_tick &&
			git commit -m "topic1: new file"
		)
	)
'

test_expect_success "edit script for multiple uploadable branches" '
	(
		cd work &&
		cat >expect<<-EOF &&
		INFO: editor is '"'"':'"'"', return directly:
		# Uncomment the branches to upload:
		#
		# project main/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		#
		# project projects/app1/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		FATAL: nothing uncommented for upload
		EOF
		test_must_fail git-repo upload \
			-v \
			--assume-no \
			--mock-no-tty \
			--mock-ssh-info-response 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			2>&1 \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: project1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project projects/app1/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		test_must_fail git-repo upload \
			-v \
			--assume-no \
			--mock-no-tty \
			--mock-ssh-info-response 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			project1 \
			2>&1 \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: projects/app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project projects/app1/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		test_must_fail git-repo upload \
			-v \
			--assume-no \
			--mock-no-tty \
			--mock-ssh-info-response 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			projects/app1 \
			2>&1 \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project projects/app1/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		(
			cd projects &&
			test_must_fail git-repo upload \
				-v \
				--assume-no \
				--mock-no-tty \
				--mock-ssh-info-response 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
				app1 \
				2>&1
		) \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: ." '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project projects/app1/ to remote branch :
		  branch my/topic1 ( 1 commit(s)):
		         <hash>
		to https://example.com (y/N)? No
		Error: upload aborted by user
		EOF
		(
			cd projects/app1 &&
			test_must_fail git-repo upload \
				-v \
				--assume-no \
				--mock-no-tty \
				--mock-ssh-info-response 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
				. \
				2>&1
		) \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: main, projects/app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		INFO: editor is '"'"':'"'"', return directly:
		# Uncomment the branches to upload:
		#
		# project main/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		#
		# project projects/app1/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		FATAL: nothing uncommented for upload
		EOF
		test_must_fail git-repo upload \
			-v \
			--assume-no \
			--mock-no-tty \
			--mock-ssh-info-response 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			main projects/app1 projects/app2 \
			2>&1 \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "upload with args: main, projects/app1" '
	(
		cd work &&
		cat >expect<<-EOF &&
		INFO: editor is '"'"':'"'"', return directly:
		# Uncomment the branches to upload:
		#
		# project main/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		#
		# project projects/app1/:
		#  branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/main.git refs/heads/my/topic1:refs/for/maint/my/topic1
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/project1.git refs/heads/my/topic1:refs/for/maint/my/topic1
		
		----------------------------------------------------------------------
		EOF
		cat >mock-edit-script<<-EOF &&
		INFO: editor is '"'"':'"'"', return directly:
		# Uncomment the branches to upload:
		#
		# project main/:
		branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		#
		# project projects/app1/:
		 branch my/topic1 ( 1 commit(s)) to remote branch maint:
		#         <hash>
		FATAL: nothing uncommented for upload
		EOF
		git-repo upload \
			-v \
			--assume-no \
			--mock-no-tty \
			--mock-git-push \
			--mock-edit-script=mock-edit-script \
			--mock-ssh-info-response 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
			main projects/app1 projects/app2 \
			2>&1 \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_expect_success "create commits" '
	(
		cd work &&
		git repo start --all my/topic1 &&
		(
			cd projects/app1 &&
			for i in $(seq 1 5)
			do
				test_tick &&
				git commit --allow-empty -m "commit #$i"
			done
		)
	)
'

test_expect_success "if has many commits, must confirm before upload" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Upload project projects/app1/ to remote branch :
		  branch my/topic1 ( 6 commit(s)):
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
		NOTE: will execute command: git push --receive-pack=agit-receive-pack ssh://git@ssh.example.com:22/project1.git refs/heads/my/topic1:refs/for/maint/my/topic1
		
		----------------------------------------------------------------------
		EOF
		(
			git-repo upload \
				-v \
				--assume-yes \
				--mock-no-tty \
				--mock-git-push \
				--mock-ssh-info-response 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}" \
				projects/app1 \
				2>&1
		) \
		| sed -e "s/[0-9a-f]\{40\}/<hash>/g" \
		>actual &&
		test_cmp expect actual
	)
'

test_done
