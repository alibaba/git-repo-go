#!/bin/sh

git_current_branch() {
	branch=$(git symbolic-ref HEAD 2>/dev/null | sed -e "s#refs/heads/##")
	if test -n "$branch"
	then
		echo "$branch"
	else
		echo "Detached HEAD"
	fi
}
