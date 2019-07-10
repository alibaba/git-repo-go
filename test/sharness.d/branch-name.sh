#!/bin/sh

git_current_branch() {
	branch=$(git branch -l | grep "^*")
	if echo "$branch" | grep -q "detached "
	then
		echo "Detached HEAD"
	else
		echo "${branch#* }"
	fi
}
