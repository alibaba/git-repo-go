# Repository

Repository is a wrapper for bare object repository or non-bare work repository.

Repository may has a reference repository, which has objects can be borrowed.
See `LocalReference` field。


# Go-Git

In order to increase the speed of `git-repo`, we try to use go-git package to
handle git repository instead of calling git command.

Call Raw() will returns raw go-git repository object, E.g.

    r := repo.Raw()
    head, err := repo.Head()


# Config for workspace

We use `goconfig` package, a pure go pckage instead of git command, to read
and write git config.

    cfg := repo.Config()
    name := cfg.Get("user.name")

Save config back to manifest project, using:

    cfg := repo.Config()

    // Change cfg settings

    err := repo.SaveConfig(cfg)


# Testing repository

To add test cases for repository, please see `project/repository_test.go`.
