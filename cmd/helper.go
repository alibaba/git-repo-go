// Copyright 2019 Hugo project

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// commandError is an error used to signal different error situations in command handling.
type commandError struct {
	s         string
	userError bool
}

func (c commandError) Error() string {
	return c.s
}

func (c commandError) isUserError() bool {
	return c.userError
}

func newUserError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: true}
}

func newUserErrorF(format string, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintf(format, a...), userError: true}
}

func newSystemError(a ...interface{}) commandError {
	return commandError{s: fmt.Sprintln(a...), userError: false}
}

func newSystemErrorF(format string, a ...interface{}) commandError {
	return commandError{s: fmt.Sprintf(format, a...), userError: false}
}

// Catch some of the obvious user errors from Cobra.
// We don't want to show the usage message for every error.
// The below may be to generic. Time will show.
var userErrorRegexp = regexp.MustCompile("argument|flag|shorthand")

func isUserError(err error) bool {
	if cErr, ok := err.(commandError); ok && cErr.isUserError() {
		return true
	}

	return userErrorRegexp.MatchString(err.Error())
}

func min(args ...int) int {
	m := args[0]
	for _, arg := range args[1:] {
		if arg < m {
			m = arg
		}
	}
	return m
}

func userInput(prompt, defaultValue string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	if text == "" {
		return defaultValue
	}
	return text
}

func answerIsTrue(answer string) bool {
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" ||
		answer == "yes" ||
		answer == "t" ||
		answer == "true" ||
		answer == "on" ||
		answer == "1" {
		return true
	}
	return false
}
