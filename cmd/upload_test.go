package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadOptionsExport(t *testing.T) {
	var (
		actual, expect string
		assert         = assert.New(t)
	)

	o := uploadOptions{
		Cc: []string{
			"知忧",
			"良久",
		},
		Reviewers: []string{
			"星楚",
			"澳明",
		},
		Title: "Hacks from 昕希",
		Description: `Detail changes:

+ Change 1 ...
+ Change 2 ...`,
		Draft: true,
		Issue: "123",
	}

	actual = strings.Join(o.Export(false), "\n")
	expect = `# [Title]       : one line message below as the title of code review

Hacks from 昕希

# [Description] : multiple lines of text as the description of code review

Detail changes:

+ Change 1 ...
+ Change 2 ...

# [Issue]       : multiple lines of issue IDs for cross references

123

# [Reviewer]    : multiple lines of user names as the reviewers for code review

星楚
澳明

# [Cc]          : multiple lines of user names as the watchers for code review

知忧
良久

# [Draft]       : a boolean (yes/no, or true/false) to turn on/off draft mode

yes

# [Private]     : a boolean (yes/no, or true/false) to turn on/off private mode
`
	assert.Equal(expect, actual)
}

func TestUploadOptionsLoad(t *testing.T) {
	var (
		data   string
		assert = assert.New(t)
		o      = uploadOptions{}
	)
	data = `# [Title]       : one line message below as the title of code review

Hacks from 昕希

# [Description] : multiple lines of text as the description of code review

Detail changes:

+ Change 1 ...
+ Change 2 ...

# [Issue]       : multiple lines of issue IDs for cross references

123

# [Reviewer]    : multiple lines of user names as the reviewers for code review

星楚
澳明

# [Cc]          : multiple lines of user names as the watchers for code review

知忧
良久

# [Draft]       : a boolean (yes/no, or true/false) to turn on/off draft mode

true

# [Private]     : a boolean (yes/no, or true/false) to turn on/off private mode
`

	o.LoadFromText(data)
	assert.Equal(o,
		uploadOptions{
			Cc: []string{
				"知忧",
				"良久",
			},
			Reviewers: []string{
				"星楚",
				"澳明",
			},
			Title: "Hacks from 昕希",
			Description: `Detail changes:

+ Change 1 ...
+ Change 2 ...`,
			Draft: true,
			Issue: "123",
		},
	)
}
