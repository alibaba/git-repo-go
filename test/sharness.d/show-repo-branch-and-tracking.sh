#!/bin/sh

show_all_repo_branch_tracking() {
	git-repo forall '
		echo "## $REPO_PATH" &&
		branch=$(git branch 2> /dev/null | sed -e "/^[^*]/d" -e "s/* \(.*\)/\1/") &&
		printf "   $branch => " &&
		(git config branch.${branch}.merge || true)
	'
}
