package editor

import (
	"testing"

	"code.alibaba-inc.com/force/git-repo/cap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedOS struct {
	mock.Mock
}

func (v *MockedOS) IsWindows() bool {
	args := v.Called()
	return args.Bool(0)
}

func TestEditorCommandsLinux(t *testing.T) {
	var (
		assert   = assert.New(t)
		mockedOS = new(MockedOS)
	)

	origCap := cap.CapWindows
	defer func() {
		cap.CapWindows = origCap
	}()

	mockedOS.On("IsWindows").Return(false)
	cap.CapWindows = mockedOS

	assert.Equal([]string{
		"vim",
		"file"},
		editorCommands("vim", "file"))
	assert.Equal([]string{
		"sh",
		"-c",
		"vim -c \"set textwidth=80\" \"$@\"",
		"sh",
		"file"},
		editorCommands("vim -c \"set textwidth=80\"", "file"))
}

func TestEditorCommandsWindows(t *testing.T) {
	var (
		assert   = assert.New(t)
		mockedOS = new(MockedOS)
	)

	origCap := cap.CapWindows
	defer func() {
		cap.CapWindows = origCap
	}()

	mockedOS.On("IsWindows").Return(true)
	cap.CapWindows = mockedOS

	assert.Equal([]string{
		"C:\\Program Files (x86)\\Notepad++\\notepad++.exe",
		"-multiInst",
		"-nosession",
		"file"},
		editorCommands("\"C:\\\\Program Files (x86)\\\\Notepad++\\\\notepad++.exe\" -multiInst -nosession",
			"file"))

	assert.Equal([]string{
		"C:\\Program Files (x86)\\Notepad++\\notepad++.exe",
		"-multiInst",
		"-nosession",
		"file"},
		editorCommands("'C:\\Program Files (x86)\\Notepad++\\notepad++.exe' -multiInst -nosession",
			"file"))

	assert.Equal([]string{
		"C:/Program Files (x86)/Notepad++/notepad++.exe",
		"-multiInst",
		"-nosession",
		"file"},
		editorCommands("'C:/Program Files (x86)/Notepad++/notepad++.exe' -multiInst -nosession",
			"file"))

}
