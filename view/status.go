package view

import termbox "github.com/nsf/termbox-go"

type StatusView struct {
	Render
	Msg string
	Win *Window
}

func NewStatusView(x, y, width, height int) *StatusView {
	return &StatusView{
		Win: newWindow(x, y, width, height),
	}
}

func (v *StatusView) Draw() {
	str := PadRight(v.Msg, v.Win.Box.Width, " ")
	tbPrint(0, v.Win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
