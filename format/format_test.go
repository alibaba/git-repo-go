package format

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsgBox(t *testing.T) {
	msgBox := NewMessageBox(10)
	pR, pW, _ := os.Pipe()
	expected := "" +
		"+----------------------------------------------------+\n" +
		"| Merge request #123 is created or updated.          |\n" +
		"|                                                    |\n" +
		"| Access MR#123 at: http://localhost/...             |\n" +
		"| Download using command: git repo download 123      |\n" +
		"+----------------------------------------------------+\n"

	msgBox.Add("Merge request #123 is created or updated.\n")
	msgBox.Add("Access MR#123 at: %s", "http://localhost/...")
	msgBox.Add("Download using command: %s", "git repo download 123")
	msgBox.SetStyle('|', '+', '-', '+', '|', '+', '-', '+')

	go func() {
		msgBox.Draw(pW)
		pW.Close()
	}()
	buf, err := ioutil.ReadAll(pR)
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, string(buf))
}

func TestMsgBoxLong(t *testing.T) {
	url := "Access MR#123 at: http://this/is/a/very/long/url, and you can also access using http://this/is/also/a/very/long/url/that/can/not/split/by/msgbox/draw/method..."

	msgBox := NewMessageBox(len(url))
	pR, pW, _ := os.Pipe()
	expected := "" +
		"###################################################################################################################################################################\n" +
		"# Merge request #123 is created or updated.                                                                                                                       #\n" +
		"#                                                                                                                                                                 #\n" +
		"# Access MR#123 at: http://this/is/a/very/long/url, and you can also access using http://this/is/also/a/very/long/url/that/can/not/split/by/msgbox/draw/method... #\n" +
		"# Download using command: git repo download 123                                                                                                                   #\n" +
		"###################################################################################################################################################################\n"

	msgBox.SetStyle('#', '#', '#', '#', '#', '#', '#', '#')
	msgBox.Add("Merge request #123 is created or updated.\n")
	msgBox.Add(url)
	msgBox.Add("Download using command: %s", "git repo download 123")

	go func() {
		msgBox.Draw(pW)
		pW.Close()
	}()
	buf, err := ioutil.ReadAll(pR)
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, string(buf))
}

func ExampleMessageBox() {
	msgBox := NewMessageBox(80)
	msgBox.Add("Merge request #123 is created or updated.\n")
	msgBox.Add("Access MR#123 at: %s", "http://localhost/...")
	msgBox.Add("Download using command: %s", "git repo download 123")
	msgBox.Draw(os.Stdout)
	// Output:
	// +----------------------------------------------------+
	// | Merge request #123 is created or updated.          |
	// |                                                    |
	// | Access MR#123 at: http://localhost/...             |
	// | Download using command: git repo download 123      |
	// +----------------------------------------------------+
}

func ExampleMessageBox_SetStyle() {
	msgBox := NewMessageBox(80)
	msgBox.Add("Merge request #123 is created or updated.\n")
	msgBox.Add("Access MR#123 at: %s", "http://localhost/...")
	msgBox.Add("Download using command: %s", "git repo download 123")
	msgBox.SetStyle('#')
	msgBox.Draw(os.Stdout)
	fmt.Println("")
	msgBox.SetStyle('|', '+', '-', '+', '|', '+', '-', '+')
	msgBox.Draw(os.Stdout)
	// Output:
	// ######################################################
	// # Merge request #123 is created or updated.          #
	// #                                                    #
	// # Access MR#123 at: http://localhost/...             #
	// # Download using command: git repo download 123      #
	// ######################################################
	//
	// +----------------------------------------------------+
	// | Merge request #123 is created or updated.          |
	// |                                                    |
	// | Access MR#123 at: http://localhost/...             |
	// | Download using command: git repo download 123      |
	// +----------------------------------------------------+
}
