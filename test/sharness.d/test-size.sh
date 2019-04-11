#!/bin/sh

# get_size <path>
get_size() {
	if test $# -ne 1
	then
		echo >&2 "Usage: get_sizeof <path>"
		return 1
	fi

	if ! test -e "$1"
	then
		return 1
	fi
	du -Lsk "$1" | awk '{print $1}'
}

# test_size <path> -gt 1000
test_size() {
	if test $# -ne 3
	then
		echo >&2 "Usage: test_size <path> <op> <number>"
		return 1
	fi

	name=$1
	size=$(get_size "$name")

	if test -z "$size"
	then
		echo >&2 "ERROR: cannot find path '$1'"
		return 1
	fi

	shift

	if ! eval "test $size $@"
	then
		echo >&2 "ERROR: size of '$name' is $size, test failed: test $size $@"
		return 1
	fi
}
