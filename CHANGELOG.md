# Changelog

Changes of git-repo.

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
