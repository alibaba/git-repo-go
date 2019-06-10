# Changelog

Changes of git-repo.

## 0.2.0 (2019-6-9)

+ "git repo --version" follows the same rule as version command
+ test: add test cases for submodule projects
+ refactor: urlJoin should keep spaces unchanged
+ Add build tag for windows build support
+ test: use git peer-review instead of git review
+ When comparing, undefined version is lower then others
+ New alias: git pr, stands for git peer-review
+ version: check if git-repo aliases can be use safely
+ Install ~/.git-repo/config.yml.example file
+ refactor: store extra git config and comments in string

## 0.1.0 (2019-6-5)

+ upgrade: show download progress
+ upgrade: add --no-cert-checks option
+ upgrade: validate package by sha256 sum and gpg signature
+ refactor: viper only bind necessary flags in rootcmd
+ bugfix: continue push if consume yes on dirty worktree

## 0.0.3 (2019-5-29)

New feature:

+ New command: git repo upgrade
+ New command: git repo status
+ New command: git-repo forall

Enhancement and bugfix:

+ Get project's HEAD from .gitdir inside worktree
+ go.mod: update modules goconfig and multi-log
+ refactor: move executeResult from forall to project
+ sync: not quit if fail to check remote server type
+ sync: return error if work repo is nil when syncing
+ goconfig: fix index out of range bug
+ upload: use knownReviewHosts to help to find review URL
+ upload: show log info if cannot upload a branch
+ upload: one dialog for options and branches edition
+ upload: use more readable upload options message
+ upload: New option --no-edit
+ Update edit script error message
+ Open an editor for user to custom upload options
+ refactor: parse reviewers later in UploadAndReport method

## 0.0.2 (2019-5-20)

+ test: add test case for install hooks
+ Link gerrit hooks when sync repo from gerrit
+ Install git-hook templates to ~/.git-repo/hooks
+ LinkManifest failed if canot find manifest file
+ refactor: use NewEmptyRepoWorkSpace for initial workspace
+ If init from a wrong url, remove and quit
+ sync: segfault: check if ws.Manifest is nil
+ upload: add debug info for upload command
+ Encode reviewers and cc using encodeString
+ test: fixed review test url
+ If ssh port is 29418, set remote type to Gerrit
+ Hide standard ssh port for SSHInfo
+ refactor: handle review URL for single git repository
+ Do not add --receive-pack option for http url when pushing
+ Check git URL using config.ParseGitURL
+ Not check `ssh_info`, if review URL is rsync protocol
+ refactor: move git address pattern to config
+ Read ReviewURL from git config remote.origin.review
+ test: mock ssh-info API when calling git-repo sync
+ sync: call `ssh_info` API and install hooks if remote is gerrit
+ test: add test cases for git review (upload --single)
+ Add alias command 'git review'

## 0.0.1 (2019-5-14)

+ Initial versoin
