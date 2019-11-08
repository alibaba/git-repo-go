# Project

Project structure "inherit" manifest.Project, and is defined as:

    type Project struct {
        manifest.Project
        ...
    }

Extra fieds for project are:

* TopDir           : root directory of repo workspace (wich has a `.repo` subdir)
* WorkDir          : working directory of the project.
* ObjectRepository : bare repository in `.repo/project-objects/<project-name>.git`, several project may share the same ObjectRepository.  And for manifest project, the ObjectRepository is nil.
* WorkRepository   : working repository in `.repo/projects/<project-path>.git`. One project has a unique WorkRepository.


# Go-Git

In order to increase the speed of `git-repo`, we try to use go-git package to
handle git repository instead of calling git command.

Call GitRepository() will returns raw go-git repository object, E.g.

    r, err := p.GitRepository()

    head, err := r.Head()


# Config for workspace

We use `goconfig` package, a pure go pckage instead of git command, to read
and write git config.

    cfg := p.Config()
    name := cfg.Get("user.name")

To save config back to manifest project, using:

    cfg := p.Config()

    // Change cfg settings

    err := p.SaveConfig(cfg)


# Manifest project

Call NewManifestProject() will return manifest project, like:

    mp := project.NewManifestProject(rootDir, manifestURL)

The returned manifest project has all fieds of `manifest.ManifestProject`.


# Common project

Call NewProject() to return common project, like:

    for _, p := range v.Manifest.AllProjects() {
        v.Projects = append(v.Projects,
                            project.NewProject(&p, v.RootDir, manifestURL))
    }

The created project has fields of project element introduced in manifest XML.


# Testing project

To add test cases for project, please see `project/project_test.go`.
