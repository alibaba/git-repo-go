# Changelog

Changes of git-repo.

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
