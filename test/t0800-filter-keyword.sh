#!/bin/sh

test_description="upload test"

. ./lib/sharness.sh

main_repo_url="file://${REPO_TEST_REPOSITORIES}/hello/main.git"

test_expect_success "setup" '
	git repo version &&
	mkdir work &&
	(
		cd work &&
		git clone $main_repo_url main &&
		cd main &&
		cat >.gitattributes <<-EOF &&
		*.md filter=keyword-subst
		EOF
		test_tick &&
		git add .gitattributes &&
		git commit -m "Set attributes to markdown" &&
		cat >test-kw-subst.md <<-EOF &&
		Test keyword subst:
		+ Author: \$Author\$,  \$LastChangedBy\$
		+ Date: \$Date\$,  \$LastChangedDate\$
		+ Rev: \$Revision\$, \$LastChangedRevision\$
		+ Id: \$Id\$
		EOF
		test_tick &&
		git add test-kw-subst.md &&
		git commit -m "Add markdown file for kw-subst test"
	)

'

test_expect_success "keyword substitude after new checkout" '
	(
		cd work/main &&
		rm test-kw-subst.md &&
		git checkout -- test-kw-subst.md &&
		cat test-kw-subst.md
	) >actual &&
	cat >expect<<-\EOF &&
	Test keyword subst:
	+ Author: $Author: A U Thor <author@example.com> $,  $LastChangedBy: A U Thor <author@example.com> $
	+ Date: $Date: 2005-04-07 22:15:13 +0000 $,  $LastChangedDate: 2005-04-07 22:15:13 +0000 $
	+ Rev: $Revision: v1.0.0-3-g0d7abb9 $, $LastChangedRevision: v1.0.0-3-g0d7abb9 $
	+ Id: $Id: test-kw-subst.md v1.0.0-3-g0d7abb9 2005-04-07 22:15:13 +0000 A U Thor <author@example.com> $
	EOF
	test_cmp expect actual
'

test_done
