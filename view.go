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

type Layer struct {
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func NewLayer(x, y, width, height int) *Layer {
	return &Layer{
		win:       newWindow(x, y, width, height),
		cursorPos: newPosition(0, 0),
		drawPos:   newPosition(0, 0),
	}
}

func (l *Layer) getCursorY() int {
	return l.win.DrawY(l.cursorPos.Y) - l.drawPos.Y
}

func (l *Layer) getDrawY(i int) int {
	return l.win.DrawY(i) - l.drawPos.Y
}

func (l *Layer) DrawBackGround(fg, bg termbox.Attribute) {
	for i := 0; i < l.win.Box.Height; i++ {
		drawStr := PadRight("", l.win.Box.Width, " ")
		drawX := l.win.DrawX(0)
		drawY := l.win.DrawY(i)
		tbPrint(drawX, drawY, fg, bg, drawStr)
	}
}

func (l *Layer) DrawContents(
	lines []string,
	cursorFG, cursorBG termbox.Attribute,
	defaultFG, defaultBG termbox.Attribute,
) {
	for i, line := range lines {
		if i >= l.drawPos.Y {
			drawStr := PadRight(line, l.win.Box.Width, " ")
			drawX := l.win.DrawX(0)
			drawY := l.getDrawY(i)
			var fg, bg termbox.Attribute
			if drawY == l.getCursorY() {
				fg = cursorFG
				bg = cursorBG
			} else {
				fg = defaultFG
				bg = defaultBG
			}
			tbPrint(drawX, drawY, fg, bg, drawStr)
		}
	}
}

func (l *Layer) UpCursor(val int) int {
	l.cursorPos.Y -= val
	if l.cursorPos.Y < 0 {
		l.cursorPos.Y = 0
	}
	if l.cursorPos.Y < l.drawPos.Y {
		l.drawPos.Y = l.cursorPos.Y
	}
	log.Printf("Up detail. CursorPosition:%d, DrawPosition:%d", l.cursorPos.Y, l.drawPos.Y)
	return l.cursorPos.Y
}

func (l *Layer) DownCursor(val int, contentNum int) int {
	l.cursorPos.Y += val
	if l.cursorPos.Y > (contentNum - 1) {
		l.cursorPos.Y = contentNum - 1
	}
	if l.cursorPos.Y > (l.drawPos.Y + l.win.Box.Height - 1) {
		l.drawPos.Y = l.cursorPos.Y - l.win.Box.Height + 1
	}
	log.Printf("Down detail. CursorPosition:%d, DrawPosition:%d", l.cursorPos.Y, l.drawPos.Y)
	return l.cursorPos.Y
}

func (l *Layer) HalfPageUpCursor() int {
	_, height := termbox.Size()
	halfPage := height / 2
	return l.UpCursor(halfPage)
}

func (l *Layer) HalfPageDownCursor(contentNum int) int {
	_, height := termbox.Size()
	halfPage := height / 2
	return l.DownCursor(halfPage, contentNum)
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
	key      string
	listType S3ListType
	objects  []*S3Object
	layer    *Layer
}

func (v *ListView) Draw() {
	for i, obj := range v.objects {
		drawStr := obj.Name
		if v.listType == ObjectList {
			drawStr = strings.TrimPrefix(obj.Name, v.key)
		}

		if i >= v.layer.drawPos.Y {
			drawY := v.layer.getDrawY(i)
			var fg, bg termbox.Attribute
			if drawY == v.layer.getCursorY() {
				drawStr = PadRight(drawStr, v.layer.win.Box.Width, " ")
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
}

func (v *ListView) getCursorObject() *S3Object {
	return v.objects[v.layer.cursorPos.Y]
}

func (v *ListView) updateList(node *Node) {
	v.layer.cursorPos.Y = node.position
	v.objects = node.objects
	v.key = node.key
	v.listType = node.GetType()
}

func (v *ListView) up() int {
	return v.layer.UpCursor(1)
}

func (v *ListView) down() int {
	return v.layer.DownCursor(1, len(v.objects))
}

func (v *ListView) halfPageUp() int {
	return v.layer.HalfPageUpCursor()
}

func (v *ListView) halfPageDown() int {
	return v.layer.HalfPageDownCursor(len(v.objects))
}

type NavigationView struct {
	Render
	currentPath string
	win         *Window
}

func (v *NavigationView) SetCurrentPath(bucket string, node *Node) {
	if node.IsRoot() {
		v.currentPath = "list bucket"
		return
	}

	showBucketName := fmt.Sprintf("s3://%s", bucket)
	if node.IsBucketRoot() {
		v.currentPath = showBucketName
	} else {
		v.currentPath = strings.Join([]string{showBucketName, node.key}, "/")
	}
}

func (v *NavigationView) Draw() {
	str := PadRight(v.currentPath, v.win.Box.Width, " ")
	tbPrint(0, v.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
