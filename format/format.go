package format

import (
	"fmt"
	"io"
	"strings"
)

// MessageBox is used to draw a message box
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

// NewMessageBox creates new message box instance
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

	// msgBox.SetStyle('#', '#', '#', '#', '#', '#', '#', '#')
	msgBox.SetStyle('|', '+', '-', '+', '|', '+', '-', '+')
	return &msgBox
}

// SetStyle changes style of message box
func (v *MessageBox) SetStyle(l, tl, t, tr, r, br, b, bl byte) {
	v.leftC = l
	v.topLeftC = tl
	v.topC = t
	v.topRightC = tr
	v.rightC = r
	v.bottomRightC = br
	v.bottomC = b
	v.bottomLeftC = bl
}

// Add messages to message box
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

// Draw shows message box
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
