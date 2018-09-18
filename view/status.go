package view

import termbox "github.com/nsf/termbox-go"

type StatusView struct {
	Render
	Msg string
	win *Window
}

func NewStatusView(x, y, width, height int) *StatusView {
	return &StatusView{
		win: newWindow(x, y, width, height),
	}
}

func (v *StatusView) Draw() {
	str := PadRight(v.Msg, v.win.Box.Width, " ")
	tbPrint(0, v.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
