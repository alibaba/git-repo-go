#!/bin/sh

test_description="git peer-review on repo with no remote"

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
	git config alias.pr >>actual &&
	cat >expect <<-EOF &&
	repo upload --single
	EOF
	test_cmp expect actual
'

test_expect_success "remove remote" '
	(
		cd work/main &&
		git remote remove origin
	)
'

test_expect_success "upload error: not remote defined" '
	(
		cd work &&
		cat >expect<<-EOF &&
		Error: no remote defined for project main
		EOF
		(
			cd main &&
			test_must_fail git pr \
				--assume-yes \
				--no-edit \
				--mock-ssh-info-status 200 \
				--mock-ssh-info-response \
				"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
		) >actual 2>&1 &&
		test_cmp expect actual
	)
'

test_done
