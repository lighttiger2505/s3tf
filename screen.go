package main

import (
	"fmt"
	"log"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Render interface {
	Draw()
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

type S3ListType int

const (
	BucketList S3ListType = iota //0
	BucketRootList
	ObjectList
)

type ListView struct {
	Render
	key       string
	listType  S3ListType
	objects   []*S3Object
	bucket    string
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (w *ListView) Draw() {
	for i, obj := range w.objects {
		drawStr := obj.Name
		if w.listType == ObjectList {
			drawStr = strings.TrimPrefix(obj.Name, w.key)
		}

		if i >= w.drawPos.Y {
			drawY := w.win.DrawY(i) - w.drawPos.Y
			var fg, bg termbox.Attribute
			if drawY == w.getCursorY() {
				drawStr = PadRight(drawStr, w.win.Box.Width, " ")
				fg = termbox.ColorWhite
				bg = termbox.ColorGreen
			} else if Bucket == obj.ObjType || PreDir == obj.ObjType || Dir == obj.ObjType {
				fg = termbox.ColorGreen
				bg = termbox.ColorDefault
			} else {
				fg = termbox.ColorDefault
				bg = termbox.ColorDefault
			}
			tbPrint(0, drawY, fg, bg, drawStr)
		}
	}

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

func (w *ListView) getCursorY() int {
	return w.win.DrawY(w.cursorPos.Y) - w.drawPos.Y
}

func (w *ListView) getCursorObject() *S3Object {
	return w.objects[w.cursorPos.Y]
}

func (w *ListView) updateList(node *Node) {
	w.cursorPos.Y = node.position
	w.objects = node.objects
	w.key = node.key
	w.listType = node.GetType()
}

func (w *ListView) up() int {
	if w.cursorPos.Y > 0 {
		w.cursorPos.Y--
	}
	if w.cursorPos.Y < w.drawPos.Y {
		w.drawPos.Y = w.cursorPos.Y
	}
	log.Printf("Up. CursorPosition:%d, DrawPosition:%d", w.cursorPos.Y, w.drawPos.Y)
	return w.cursorPos.Y
}

func (w *ListView) down() int {
	if w.cursorPos.Y < (len(w.objects) - 1) {
		w.cursorPos.Y++
	}
	if w.cursorPos.Y > (w.drawPos.Y + w.win.Box.Height - 1) {
		w.drawPos.Y = w.cursorPos.Y - w.win.Box.Height + 1
	}
	log.Printf("Down. CursorPosition:%d, DrawPosition:%d", w.cursorPos.Y, w.drawPos.Y)
	return w.cursorPos.Y
}

type NavigationView struct {
	Render
	key string
	win *Window
}

func (w *NavigationView) SetKey(bucket, key string) {
	if key == "" {
		w.key = "list bucket"
	} else if bucket == key {
		w.key = bucket
	} else {
		w.key = strings.Join([]string{bucket, key}, "/")
	}
}

func (w *NavigationView) Draw() {
	str := PadRight(w.key, w.win.Box.Width, " ")
	tbPrint(0, w.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}

type StatusView struct {
	Render
	msg string
	win *Window
}

func (w *StatusView) Draw() {
	str := PadRight(w.msg, w.win.Box.Width, " ")
	tbPrint(0, w.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
