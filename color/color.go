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

// Package color implements color output on console.
package color

import (
	"fmt"

	"github.com/aliyun/git-repo-go/cap"
)

type mapping map[string]int

func (v mapping) get(s string) int {
	num, ok := v[s]
	if !ok {
		return -1
	}
	return num
}

func (v mapping) has(s string) bool {
	_, ok := v[s]
	return ok
}

const (
	colorPrefix = "\033["
	colorReset  = "\033[m"
)

var colorMap = mapping{
	"normal":  -1,
	"black":   0,
	"red":     1,
	"green":   2,
	"yellow":  3,
	"blue":    4,
	"magenta": 5,
	"cyan":    6,
	"white":   7,
}

var attrMap = mapping{
	"bold":    1,
	"dim":     2,
	"ul":      4,
	"blink":   5,
	"reverse": 7,
}

func colorEnabled() bool {
	if !cap.Isatty() || cap.IsWindows() {
		return false
	}
	return true
}

// Color returns color code for terminal display
//
// Available colors:
//	normal
//	black
//	red
//	green
//	yellow
//	blue
//	magenta
//	cyan
//	white
//
// Available attributes:
//	bold
//	dim
//	ul
//	blink
//	reverse
func Color(fgColor, bgColor, attrVal string) string {
	var (
		code string
	)

	if !colorEnabled() {
		return ""
	}

	fg := colorMap.get(fgColor)
	bg := colorMap.get(bgColor)
	attr := attrMap.get(attrVal)

	if attr >= 0 || fg >= 0 || bg >= 0 {
		if attr >= 0 {
			code = string('0' + attr)
		}

		if fg >= 0 {
			if len(code) > 0 {
				code += ";"
			}

			if fg < 8 {
				code += "3" + string('0'+fg)
			} else {
				code += "38;5;" + string('0'+fg)
			}
		}

		if bg >= 0 {
			if len(code) > 0 {
				code += ";"
			}
			if bg < 8 {
				code += "4" + string('0'+bg)
			} else {
				code += "48;5;" + string('0'+bg)
			}
		}
		code += "m"
	}
	if len(code) > 0 {
		return colorPrefix + code
	}
	return ""
}

// Reset returns ResetColor to reset color output.
func Reset() string {
	if colorEnabled() {
		return colorReset
	}
	return ""
}

// Hilight shows hightlight message
func Hilight(msg string) {
	fmt.Printf("%s%s%s",
		Color("", "", "bold"),
		msg,
		Reset(),
	)
}

// Hilightln shows hightlight message with LF
func Hilightln(msg string) {
	fmt.Printf("%s%s%s\n",
		Color("", "", "bold"),
		msg,
		Reset(),
	)
}

// Hilightf is sprintf version of Hilight
func Hilightf(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	Hilight(msg)
}

// Dim shows message in dim style
func Dim(msg string) {
	fmt.Printf("%s%s%s",
		Color("", "", "dim"),
		msg,
		Reset(),
	)
}

// Dimln shows message in dim style with LF
func Dimln(msg string) {
	fmt.Printf("%s%s%s\n",
		Color("", "", "dim"),
		msg,
		Reset(),
	)
}

// Dimf is sprintf version of Dim
func Dimf(f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	Dim(msg)
}
