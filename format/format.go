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

// Package format implements our own content print methods.
package format

import (
	"fmt"
	"io"
	"strings"
)

// MessageBox defines message box structure used to draw pretty message box.
type MessageBox struct {
	msgs     []string
	maxWidth int
	width    int

	leftC        byte
	rightC       byte
	topC         byte
	bottomC      byte
	topLeftC     byte
	topRightC    byte
	bottomLeftC  byte
	bottomRightC byte
}

// NewMessageBox creates new message box instance.
func NewMessageBox(maxWidth int) *MessageBox {
	width := 50
	if width > maxWidth {
		maxWidth = width
	}

	msgBox := MessageBox{
		msgs:         []string{},
		width:        width,
		maxWidth:     maxWidth,
		leftC:        '|',
		rightC:       '|',
		topC:         '-',
		bottomC:      '-',
		topLeftC:     '+',
		topRightC:    '+',
		bottomLeftC:  '+',
		bottomRightC: '+',
	}

	return &msgBox
}

// SetStyle changes style of message box. Arguments provide position code for style.
//
// One position code: draw message box border using the given character.
//
// Four position codes: draw message box border using the given characters.
// The 1st code defines the left border, the 2nd code defines the top border,
// The 3rd code defines the right border, and the last code defines the bottom
// border.
//
// Eight position codes: draw message box border using the given characters.
// Chacters given by the position codes defiens the left, top-left, top,
// top-right, right, bottom-right, bottom, bottom-left charcter of the border.
func (v *MessageBox) SetStyle(pos ...byte) {
	var (
		l, tl, t, tr, r, br, b, bl byte
	)

	if len(pos) == 1 {
		l, tl, t, tr, r, br, b, bl =
			pos[0],
			pos[0],
			pos[0],
			pos[0],
			pos[0],
			pos[0],
			pos[0],
			pos[0]
	} else if len(pos) == 4 {
		l, tl, t, tr, r, br, b, bl =
			pos[0],
			pos[1],
			pos[1],
			pos[1],
			pos[2],
			pos[3],
			pos[3],
			pos[3]
	} else if len(pos) == 8 {
		l, tl, t, tr, r, br, b, bl =
			pos[0],
			pos[1],
			pos[2],
			pos[3],
			pos[4],
			pos[5],
			pos[6],
			pos[7]
	} else {
		l, tl, t, tr, r, br, b, bl =
			'*',
			'*',
			'*',
			'*',
			'*',
			'*',
			'*',
			'*'
	}

	v.leftC = l
	v.topLeftC = tl
	v.topC = t
	v.topRightC = tr
	v.rightC = r
	v.bottomRightC = br
	v.bottomC = b
	v.bottomLeftC = bl
}

// Add messages to message box.
func (v *MessageBox) Add(a ...interface{}) {
	var msg string
	if len(a) == 1 {
		msg = a[0].(string)
	} else if len(a) > 1 {
		msg = fmt.Sprintf(a[0].(string), a[1:]...)
	}
	for _, line := range strings.Split(msg, "\n") {
		line = strings.TrimRight(line, " \t\r")
		v.addLine(line)
	}
}

func (v *MessageBox) addLine(line string) {
	for len(line) > v.maxWidth {
		i := v.maxWidth - 1
		for ; ; i-- {
			if line[i] == ' ' {
				v.msgs = append(v.msgs, line[0:i])
				line = line[i+1:]
				if v.width < i {
					v.width = i
				}
				break
			}
			// stop word not find in line
			if i <= v.maxWidth/2+1 {
				v.msgs = append(v.msgs, line[0:v.maxWidth])
				v.width = v.maxWidth
				line = line[v.maxWidth:]
				break
			}
		}
	}
	if v.width < len(line) {
		v.width = len(line)
	}
	v.msgs = append(v.msgs, line)
}

// Draw starts to draw message box.
func (v *MessageBox) Draw(w io.Writer) {
	fmt.Fprintf(w, "%c%s%c\n",
		v.topLeftC,
		strings.Repeat(string(v.topC), v.width+2),
		v.topRightC)
	for _, line := range v.msgs {
		fmt.Fprintf(w, "%c %-*s %c\n",
			v.leftC,
			v.width, line,
			v.rightC)
	}
	fmt.Fprintf(w, "%c%s%c\n",
		v.bottomLeftC,
		strings.Repeat(string(v.bottomC), v.width+2),
		v.bottomRightC)
}
