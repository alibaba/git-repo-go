package helper

import (
	"bytes"
	"strconv"
	"strings"
)

// ReplaceMacros replaces macros in string.
func ReplaceMacros(s string, macros map[string]string) string {
	var (
		ret strings.Builder
		err error
	)

	p := []byte(s)
	for {
		var (
			key, mod    string
			value       []byte
			width       int
			left, right int
			loc         int
		)

		left = bytes.IndexAny(p, "{<")
		if left == -1 {
			ret.Write(p)
			break
		}
		right = bytes.IndexAny(p, "}>")
		if right == -1 {
			ret.Write(p)
			break
		}
		if right < left {
			ret.Write(p[:right+1])
			p = p[right+1:]
			continue
		}
		ret.Write(p[:left])
		field := p[left+1 : right]
		loc = bytes.IndexByte(field, ':')
		key = string(field)
		if loc > 0 {
			key = string(field[:loc])
			field = field[loc+1:]
			loc = bytes.IndexByte(field, ':')
			if loc > 0 {
				mod = string(field[:loc])
				width, err = strconv.Atoi(string(field[loc+1:]))
				if err != nil {
					ret.Write(p[left : right+1])
					p = p[right+1:]
					continue
				}
			}
		}
		match := false
		for macroKey, macroValue := range macros {
			if key == macroKey {
				value = []byte(macroValue)
				match = true
				break
			}
		}
		if !match {
			ret.Write(p[left : right+1])
			p = p[right+1:]
			continue
		}
		if len(mod) > 0 {
			if mod == "left" {
				if width <= len(value) {
					value = value[:width]
				} else {
					newValue := bytes.Repeat([]byte("0"), width-len(value))
					newValue = append(newValue, value...)
					value = newValue
				}
			} else if mod == "right" {
				if width <= len(value) {
					value = value[len(value)-width:]
				} else {
					newValue := bytes.Repeat([]byte("0"), width-len(value))
					newValue = append(newValue, value...)
					value = newValue
				}
			} else {
				ret.Write(p[left : right+1])
				p = p[right+1:]
				continue
			}
		}
		ret.Write(value)
		p = p[right+1:]
	}
	return ret.String()
}
