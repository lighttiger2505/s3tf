package main

import (
	"fmt"
	"log"

	termbox "github.com/nsf/termbox-go"
)

type Render interface {
	Draw()
}

type EventHandler interface {
	Handle(termbox.Event)
}

type Position struct {
	X, Y int
}

func newPosition(x, y int) *Position {
	return &Position{x, y}
}

type Box struct {
	Width, Height int
}

type Window struct {
	Pos *Position
	Box *Box
}

func newWindow(x, y, width, height int) *Window {
	return &Window{
		Pos: &Position{
			X: x,
			Y: y,
		},
		Box: &Box{
			Width:  width,
			Height: height,
		},
	}
}

func (w *Window) DrawX(x int) int {
	return w.Pos.X + x
}

func (w *Window) DrawY(y int) int {
	return w.Pos.Y + y
}

type BucketList struct {
	Render
	EventHandler
	buckets   []string
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (w *BucketList) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	for i, bucket := range w.buckets {
		// if i >= w.drawPos.Y && i <= (w.drawPos.Y+w.win.Box.Height) {
		if i >= w.drawPos.Y {
			tbPrint(0, w.win.DrawY(i)-w.drawPos.Y, termbox.ColorDefault, termbox.ColorDefault, bucket)
		}
	}
	termbox.SetCursor(0, w.win.DrawY(w.cursorPos.Y)-w.drawPos.Y)

	status := fmt.Sprintf(
		"pos: (%d, %d) draw: (%d, %d) box: (%d, %d)",
		w.cursorPos.X,
		w.cursorPos.Y,
		w.drawPos.X,
		w.drawPos.Y,
		w.win.Box.Width,
		w.win.Box.Height,
	)
	log.Println(status)
}

func (w *BucketList) Handle(ev termbox.Event) {
	if ev.Ch == 'j' {
		if w.cursorPos.Y < (len(w.buckets) - 1) {
			w.cursorPos.Y++
		}
		if w.cursorPos.Y > (w.drawPos.Y + w.win.Box.Height - 1) {
			w.drawPos.Y = w.cursorPos.Y - w.win.Box.Height + 1
		}
	} else if ev.Ch == 'k' {
		if w.cursorPos.Y > 0 {
			w.cursorPos.Y--
		}
		if w.cursorPos.Y < w.drawPos.Y {
			w.drawPos.Y = w.cursorPos.Y
		}
	}
}
