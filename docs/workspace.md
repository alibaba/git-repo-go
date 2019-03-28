# Workspace

Workspace is the toplevel structure/interface for projects in current
`git-repo` workspace.

For multple projects managed by manifest XML, workspace objects contains
manifest project and all other projects introduced by the manifest XML.

TODO: If the option `--single` is provided, workspace is only a wrapper
for the current repository.


# Find and return repo workspace

Call `NewWorkSpace(<dir>)` will start to detect repo workspace from `<dir>`
(default is current working dir).

    ws, err := workspace.NewWorkSpace("")

It will:

1. Searching a hidden `.repo` directory in `<dir>` or any parent directory. 
2. Returns a WorkSpace objects based on the toplevel directory of workspace.
3. If cannot find valid repo workspace, return ErrRepoDirNotFound error.

For command `git repo init`, it should ignore ErrRepoDirNotFound error should
call:

    ws, err := workspace.NewWorkSpaceInit("", initOptions.ManifestURL)

NewWorkSpaceInit will also try to search workspace root, but if is not found,
will use `<dir>` as root of a new workspace.


# Config for workspace

Configurations of workspace are saved in git config of manifest repository.

Read config using:

    cfg := ws.Config()
    name := cfg.Get("manifest.name")

Save config back to manifest proejct, using:

    cfg := ws.Config()

    // Change cfg settings

    err := ws.SaveConfig(cfg)


# ManifestURL changed, and reset URL of all projects in workspace

Call Load() to read manifest XML file and reset ManifestURL if it changed,
and reset URL of all projects in workspace.

    err := ws.Load(manifestURL)


# Testing workspace

Test cases for workspace, please see `workspace/workspace_test.go`.
