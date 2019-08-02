// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cap implements inspections of OS capabilities.
package cap

import (
	"os"
	"runtime"

	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/version"
	"github.com/mattn/go-isatty"
)

// WindowsInterface is the interface to implement IsWindows(),
// which checks whether current OS is Windows.
type WindowsInterface interface {
	IsWindows() bool
}

// TTYInterface is the interface to implement Isatty(),
// which checks whether a valid terminal is attached.
type TTYInterface interface {
	Isatty() bool
}

// SymlinkInterface is the interface to implement CanSymink(),
// which indicates symlink is available on current OS.
type SymlinkInterface interface {
	CanSymlink() bool
}

// GitInterface is the interface to implement Git related capabilities.
type GitInterface interface {
	GitCanPushOptions() bool
}

// Instance of interface, which can be overridden for test by mocking.
var (
	CapWindows WindowsInterface
	CapTTY     TTYInterface
	CapSymlink SymlinkInterface
	CapGit     GitInterface
)

// defaultWindowsImpl implements WindowsInterface.
type defaultWindowsImpl struct {
}

// IsWindows indicates whether current OS is windows.
func (v defaultWindowsImpl) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// defaultTTYImpl implements TTYInterface.
type defaultTTYImpl struct {
}

// Isatty indicates whether program has a valid terminal attached.
func (v defaultTTYImpl) Isatty() bool {
	if config.MockNoTTY() {
		return false
	}
	if isatty.IsTerminal(os.Stdin.Fd()) &&
		isatty.IsTerminal(os.Stdout.Fd()) {
		return true
	} else if isatty.IsCygwinTerminal(os.Stdin.Fd()) &&
		isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return true
	}

	return false
}

// defaultSymlinkImpl implements SymlinkInterface.
type defaultSymlinkImpl struct {
}

// CanSymlink indicates whether symlink is available for current OS.
func (v defaultSymlinkImpl) CanSymlink() bool {
	if config.MockNoSymlink() {
		return false
	}
	return runtime.GOOS != "windows"
}

// defaultCapGitImpl implements GitInterface.
type defaultCapGitImpl struct {
}

// GitCanPushOption indicates git can sent push-optoins or not
func (v defaultCapGitImpl) GitCanPushOptions() bool {
	return version.CompareVersion(version.GitVersion, "2.10.0") >= 0
}

// IsWindows indicates whether current OS is windows.
func IsWindows() bool {
	return CapWindows.IsWindows()
}

// CanSymlink indicates whether symlink is available for current OS.
func CanSymlink() bool {
	return CapSymlink.CanSymlink()
}

// Isatty indicates whether program has a valid terminal attached.
func Isatty() bool {
	return CapTTY.Isatty()
}

// GitCanPushOptions indicates whether git can sent push options.
func GitCanPushOptions() bool {
	return CapGit.GitCanPushOptions()
}

func init() {
	CapWindows = &defaultWindowsImpl{}
	CapTTY = &defaultTTYImpl{}
	CapSymlink = &defaultSymlinkImpl{}
	CapGit = &defaultCapGitImpl{}
}
