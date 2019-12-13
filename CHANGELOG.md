# Changelog

Changes of git-repo.

## 0.7.0 (2019-12-13)

New features:

* Add helper for protocol extension, and use can add external helper.
* Smart selection of remote if repository has many remotes defined.
* AGit-Flow 2.0: set `AGIT_FLOW` envirionment for git connection.
* AGit-Flow 2.0: support version of `ssh_info` response.
* AGit-Flow 2.0: multiple users working on one code review.
* AGit-Flow 2.0: force push prevention if oldoid is provided.

Refactors, Enhancements and bugfixes:

* Many refactors, such as project.Remote.

## 0.6.1 (2019-11-10)

+ manifest: repo compatible parsing for project revision

## 0.6.0 (2019-11-08)

NOTE: All users are suggested to upgrade to this version. Local changed files
      will not be overwrited.

+ Show project path in log message as prompt
+ init: force create default branch for manifest project
+ sync: donot overwrite local changed files
+ sync: sync -d: always detach even if nothing changed
+ sync: should not make manifest project detached
+ init: no need to set color if color.ui defined in git global/system config
+ by setting app.git.repo.disabled can git-repo for certain workspace
+ init: use git-init command to create repository

## 0.5.1 (2019-09-09)

New features:

+ upload: cache different upload options settings for different target branch
+ upload: only show title and description in editor for upload for the 1st time
+ Some commands work for repo in mirror mode

Refactors, Enhancements and bugfixes:

+ Disable upx, because some Mac users report errors
+ refactor: rename variables name, such as RepoRoot
+ refactor: Add DotGit, SharedGitDir in repository

## 0.5.0 (2019-08-20)

New features:

+ New cmd: git repo abandon
+ New cmd: git repo prune
+ New command: git repo manifest
+ cmd/manifest: freeze manifest revision if provided -r option

Refactors, Enhancements and bugfixes:

+ color: add Hilight and Dim methods
+ refactor: donot check Remote type to get reviewable branch
+ repository: get last modified of a revision
+ bugfix: not change Revision during network-half
+ refactor: make WorkRepository as embedding struction for Project
+ go.mod: update goconfig, check cache against file size
+ refactor: rename Path field name of Repository to RepoDir
+ refactor: remove ObjectRepository from project
+ refactor: IsClean only returns one boolean
+ test: add test cases for cmd/manifest
+ test: update test cases for manifest
+ refactor: rename command executable entrance name
+ test: remove pipes, which suppress errors being report


## 0.4.2 (2019-08-08)

+ Compress binaries using UPX


## 0.4.1 (2019-08-07)

+ peer-review: support --remote and --dest option
+ New command: git repo list
+ download: support --remote option
+ compatibility: Use push options only if git is greater than 2.10.0
+ compatibility: Show compatible issues of git versions
+ compatibility: Fix compatible issue of gerrit hook
+ compatibility: Use absolute path for 'include.path' directive
+ README: add badge for CI build status
+ compatibility: Use strings.Replace to be compatible with go 1.11
+ compatibility: enable GO111MODULE for build and test
+ compatibility: remove test case which depends on higher version of git
+ doc: add godoc
+ download: cherry-pick all commits for one code review


## 0.3.1 (2019-6-26)

Enhancement and bugfix:

+ bugfix: add protection for write git extra config file
+ Only set push.default to nothing if it is unset
+ Makfile: build with vendor and new release target
+ Ignore vendor dir
+ filter: ignore errors for smudge
+ debug: add more debug info for repo sync
+ upload: clean published refs for single mode


## 0.3.0 (2019-6-22) DEPRECATED

New Features:

+ Add alias command: git download
+ New command: download, for offline code review
+ Add --no-cache option to ignore `ssh_info` API cache
+ config: add new filter driver keyword-subst
+ New command: filter, for keyword-subst content filter
+ Install gerrit hooks if review server is gerrit
+ Set push.default to nothing if remote is reviewable

Enhancement and bugfix:

+ test: change branch name to upper case for test
+ goconfig: fix upper case section name issue
+ refactor LoadRemote for single repository workspace
+ ParseGitURL can parse file:// and other protocol
+ test: add mock options for git-repo sync command
+ test: add test cases for git-repo filter
+ sync: default use 4 jobs
+ test: add test case for git pr --br <branch>
+ Not quit immediately if cannot get review url
+ refactor: delay load remote for GitWorkspace
+ Only save config for DisableDefaultPush when necessary
+ Format every multi-log imports by adding alias log
+ Fix some spellings


## 0.2.1 (2019-6-26)

Enhancement and bugfix:

+ bugfix: add protection for write git extra config file
+ Makfile: build with vendor and new release target


## 0.2.0 (2019-6-9) DEPRECATED

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
