#!/bin/sh

# Setup a ".repo" file to stop search for .repo dir upward.
touch "${SHARNESS_TEST_SRCDIR}/.repo"

# Create toplevel gitdir to prevent dangous operation on current repo.
git init "$SHARNESS_TRASH_DIRECTORY"
