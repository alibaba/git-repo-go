#!/bin/sh

test_description="test 'git-repo list'"

. ./lib/sharness.sh

# Create manifest repositories
manifest_url="file://${REPO_TEST_REPOSITORIES}/hello/manifests"

test_expect_success "setup" '
	# create .repo file as a barrier, not find .repo deeper
	touch .repo &&
	mkdir work &&
	(
		cd work &&
		git-repo init -g all -u $manifest_url &&
		git-repo sync \
			--mock-ssh-info-status 200 \
			--mock-ssh-info-response \
			"{\"host\":\"ssh.example.com\", \"port\":22, \"type\":\"agit\"}"
	)
'

test_expect_success "git-repo list" '
	(
		cd work &&
		git-repo list
	) >actual &&
	cat >expect<<-EOF &&
	drivers/driver-1 : drivers/driver1
	drivers/driver-2 : drivers/driver2
	main : main
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n" '
	(
		cd work &&
		git-repo list -n
	) >actual &&
	cat >expect<<-EOF &&
	drivers/driver1
	drivers/driver2
	main
	project1
	project1/module1
	project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p" '
	(
		cd work &&
		git-repo list -p
	) >actual &&
	cat >expect<<-EOF &&
	drivers/driver-1
	drivers/driver-2
	main
	projects/app1
	projects/app1/module1
	projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f" '
	(
		cd work &&
		git-repo list -f
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/drivers/driver-1 : drivers/driver1
	.../work/drivers/driver-2 : drivers/driver2
	.../work/main : main
	.../work/projects/app1 : project1
	.../work/projects/app1/module1 : project1/module1
	.../work/projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -n" '
	(
		cd work &&
		test_must_fail git-repo list -f -n
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -f and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -n" '
	(
		cd work &&
		test_must_fail git-repo list -n -p
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -p" '
	(
		cd work &&
		git-repo list -f -p
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/drivers/driver-1
	.../work/drivers/driver-2
	.../work/main
	.../work/projects/app1
	.../work/projects/app1/module1
	.../work/projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -n -p" '
	(
		cd work &&
		test_must_fail git-repo list -f -n -p
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g app" '
	(
		cd work &&
		git-repo list -g app
	) >actual &&
	cat >expect<<-EOF &&
	main : main
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g drivers" '
	(
		cd work &&
		git-repo list -g drivers
	) >actual &&
	cat >expect<<-EOF &&
	drivers/driver-1 : drivers/driver1
	drivers/driver-2 : drivers/driver2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g app,drivers" '
	(
		cd work &&
		git-repo list -g app,drivers
	) >actual &&
	cat >expect<<-EOF &&
	drivers/driver-1 : drivers/driver1
	drivers/driver-2 : drivers/driver2
	main : main
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -g app" '
	(
		cd work &&
		git-repo list -p -g app
	) >actual &&
	cat >expect<<-EOF &&
	main
	projects/app1
	projects/app1/module1
	projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -g app" '
	(
		cd work &&
		git-repo list -n -g app
	) >actual &&
	cat >expect<<-EOF &&
	main
	project1
	project1/module1
	project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -g app" '
	(
		cd work &&
		git-repo list -f -g app
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/main : main
	.../work/projects/app1 : project1
	.../work/projects/app1/module1 : project1/module1
	.../work/projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -p -g app" '
	(
		cd work &&
		test_must_fail git-repo list -n -p -g app
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -f -g app" "
	(
		cd work &&
		test_must_fail git-repo list -n -f -g app
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -f and -n
	EOF
	test_cmp expect actual
"

test_expect_success "git-repo list -p -f -g app" '
	(
		cd work &&
		git-repo list -p -f -g app
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/main
	.../work/projects/app1
	.../work/projects/app1/module1
	.../work/projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -p -f -g app" '
	(
		cd work &&
		test_must_fail git-repo list -n -p -f -g app
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -r pro*" '
	(
		cd work &&
		git-repo list -r "pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -r ^project" '
	(
		cd work &&
		git-repo list -r "^project"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -r ^project1$" '
	(
		cd work &&
		git-repo list -r "^project[12]$"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1 : project1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -r ^pro.*" '
	(
		cd work &&
		git-repo list -n -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	project1
	project1/module1
	project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -r ^pro.*" '
	(
		cd work &&
		git-repo list -p -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1
	projects/app1/module1
	projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -r ^pro.*" '
	(
		cd work &&
		git-repo list -f -r "^pro.*"
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/projects/app1 : project1
	.../work/projects/app1/module1 : project1/module1
	.../work/projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -n -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -f -n -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -f and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -n -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -n -p -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -p -r ^pro.*" '
	(
		cd work &&
		git-repo list -f -p -r "^pro.*"
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/projects/app1
	.../work/projects/app1/module1
	.../work/projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -n -p -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -f -n -p -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g app -r ^pro.*" '
	(
		cd work &&
		git-repo list -g app -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g drivers -r ^pro.*" '
	(
		cd work &&
		git-repo list -g drivers -r "^pro.*"
	) >actual && 
	cat >expect<<-EOF &&
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -g app,drivers -r ^pro.*" '
	(
		cd work &&
		git-repo list -g app,drivers -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1 : project1
	projects/app1/module1 : project1/module1
	projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -g app -r ^pro.*" '
	(
		cd work &&
		git-repo list -p -g app -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	projects/app1
	projects/app1/module1
	projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -g app -r ^pro.*" '
	(
		cd work &&
		git-repo list -n -g app -r "^pro.*"
	) >actual &&
	cat >expect<<-EOF &&
	project1
	project1/module1
	project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -f -g app -r ^pro.*" '
	(
		cd work &&
		git-repo list -f -g app -r "^pro.*"
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/projects/app1 : project1
	.../work/projects/app1/module1 : project1/module1
	.../work/projects/app2 : project2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -p -g app -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -n -p -g app -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -f -g app -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -n -f -g app -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -f and -n
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -p -f -g app -r ^pro.*" '
	(
		cd work &&
		git-repo list -p -f -g app -r "^pro.*"
	) >out &&
	sed -e "s/^.*trash directory.t1000-list/.../g" out >actual &&
	cat >expect<<-EOF &&
	.../work/projects/app1
	.../work/projects/app1/module1
	.../work/projects/app2
	EOF
	test_cmp expect actual
'

test_expect_success "git-repo list -n -p -f -g app -r ^pro.*" '
	(
		cd work &&
		test_must_fail git-repo list -n -p -f -g app -r "^pro.*"
	) >actual 2>&1 &&
	cat >expect<<-EOF &&
	FATAL: cannot combine -p and -n
	EOF
	test_cmp expect actual
'

test_done
