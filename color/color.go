package color

import (
	"code.alibaba-inc.com/force/git-repo/cap"
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
	// Prefix is control sequences to start color output
	Prefix = "\033["
	// Reset color
	Reset = "\033[m"
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
	if !cap.Isatty() {
		return false
	}
	return true
}

func ColorReset() string {
	if colorEnabled() {
		return Reset
	}
	return ""
}

// Color returns color code for terminal display
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
		return Prefix + code
	}
	return ""
}
