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
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (v *ListView) Draw() {
	for i, obj := range v.objects {
		drawStr := obj.Name
		if v.listType == ObjectList {
			drawStr = strings.TrimPrefix(obj.Name, v.key)
		}

		if i >= v.drawPos.Y {
			drawY := v.win.DrawY(i) - v.drawPos.Y
			var fg, bg termbox.Attribute
			if drawY == v.getCursorY() {
				drawStr = PadRight(drawStr, v.win.Box.Width, " ")
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

func (v *ListView) getCursorY() int {
	return v.win.DrawY(v.cursorPos.Y) - v.drawPos.Y
}

func (v *ListView) getCursorObject() *S3Object {
	return v.objects[v.cursorPos.Y]
}

func (v *ListView) updateList(node *Node) {
	v.cursorPos.Y = node.position
	v.objects = node.objects
	v.key = node.key
	v.listType = node.GetType()
}

func (v *ListView) up() int {
	if v.cursorPos.Y > 0 {
		v.cursorPos.Y--
	}
	if v.cursorPos.Y < v.drawPos.Y {
		v.drawPos.Y = v.cursorPos.Y
	}
	log.Printf("Up. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

func (v *ListView) down() int {
	if v.cursorPos.Y < (len(v.objects) - 1) {
		v.cursorPos.Y++
	}
	if v.cursorPos.Y > (v.drawPos.Y + v.win.Box.Height - 1) {
		v.drawPos.Y = v.cursorPos.Y - v.win.Box.Height + 1
	}
	log.Printf("Down. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

type NavigationView struct {
	Render
	key string
	win *Window
}

func (v *NavigationView) SetKey(bucket, key string) {
	if key == "" {
		v.key = "list bucket"
	} else if bucket == key {
		v.key = bucket
	} else {
		v.key = strings.Join([]string{bucket, key}, "/")
	}
}

func (v *NavigationView) Draw() {
	str := PadRight(v.key, v.win.Box.Width, " ")
	tbPrint(0, v.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}

type StatusView struct {
	Render
	msg string
	win *Window
}

func (v *StatusView) Draw() {
	str := PadRight(v.msg, v.win.Box.Width, " ")
	tbPrint(0, v.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}

type MenuCommand int

const (
	CommandDownload MenuCommand = iota //0
	CommandOpen
	CommandEdit
)

type MenuItem struct {
	name      string
	shorthand string
	detail    string
	command   MenuCommand
}

func NewMenuItem(name, shorthand, detail string, command MenuCommand) *MenuItem {
	return &MenuItem{
		name:      name,
		shorthand: shorthand,
		detail:    detail,
		command:   command,
	}
}

func (i *MenuItem) toString() string {
	return fmt.Sprintf("(%s)%s %s", i.shorthand, i.name, i.detail)
}

type MenuView struct {
	Render
	items     []*MenuItem
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (v *MenuView) Draw() {
	// Draw backgroud color
	for i := 0; i < v.win.Box.Height; i++ {
		drawStr := PadRight("", v.win.Box.Width, " ")
		drawY := v.win.DrawY(i) - v.drawPos.Y
		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		tbPrint(0, drawY, fg, bg, drawStr)
	}
	// Draw menu command item
	for i, item := range v.items {
		if i >= v.drawPos.Y {
			drawStr := PadRight(item.toString(), v.win.Box.Width, " ")
			drawY := v.win.DrawY(i) - v.drawPos.Y
			var fg, bg termbox.Attribute
			if drawY == v.getCursorY() {
				fg = termbox.ColorWhite
				bg = termbox.ColorGreen
			} else {
				fg = termbox.ColorDefault
				bg = termbox.ColorDefault
			}
			tbPrint(0, drawY, fg, bg, drawStr)
		}
	}
}

func (v *MenuView) getCursorY() int {
	return v.win.DrawY(v.cursorPos.Y) - v.drawPos.Y
}

func (v *MenuView) getCursorItem() *MenuItem {
	return v.items[v.cursorPos.Y]
}

func (v *MenuView) up() int {
	if v.cursorPos.Y > 0 {
		v.cursorPos.Y--
	}
	if v.cursorPos.Y < v.drawPos.Y {
		v.drawPos.Y = v.cursorPos.Y
	}
	log.Printf("Up. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

func (v *MenuView) down() int {
	if v.cursorPos.Y < (len(v.items) - 1) {
		v.cursorPos.Y++
	}
	if v.cursorPos.Y > (v.drawPos.Y + v.win.Box.Height - 1) {
		v.drawPos.Y = v.cursorPos.Y - v.win.Box.Height + 1
	}
	log.Printf("Down. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

type DetailView struct {
	Render
	object    *S3Object
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (v *DetailView) getContents() []string {
	contents := `hogetitle

hogecontent1
hogecontent2
hogecontent3
hogecontent4
hogecontent5
hogecontent6
hogecontent7
hogecontent8
hogecontent9
hogecontent10
`
	return strings.Split(contents, "\n")
}

func (v *DetailView) getCursorY() int {
	log.Printf("getCursorY. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.win.DrawY(v.cursorPos.Y) - v.drawPos.Y
}

func (v *DetailView) up() int {
	if v.cursorPos.Y > 0 {
		v.cursorPos.Y--
	}
	if v.cursorPos.Y < v.drawPos.Y {
		v.drawPos.Y = v.cursorPos.Y
	}
	log.Printf("Up detail. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

func (v *DetailView) down() int {
	lines := v.getContents()
	if v.cursorPos.Y < (len(lines) - 1) {
		v.cursorPos.Y++
	}
	if v.cursorPos.Y > (v.drawPos.Y + v.win.Box.Height - 1) {
		v.drawPos.Y = v.cursorPos.Y - v.win.Box.Height + 1
	}
	log.Printf("Down detail. CursorPosition:%d, DrawPosition:%d", v.cursorPos.Y, v.drawPos.Y)
	return v.cursorPos.Y
}

func (v *DetailView) Draw() {
	// Draw backgroud color
	for i := 0; i < v.win.Box.Height; i++ {
		drawStr := PadRight("", v.win.Box.Width, " ")
		drawX := v.win.DrawX(0)
		drawY := v.win.DrawY(i) - v.drawPos.Y
		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		tbPrint(drawX, drawY, fg, bg, drawStr)
	}
	// Draw menu command item
	lines := v.getContents()
	for i, line := range lines {
		if i >= v.drawPos.Y {
			drawStr := PadRight(line, v.win.Box.Width, " ")
			drawX := v.win.DrawX(0)
			drawY := v.win.DrawY(i) - v.drawPos.Y
			var fg, bg termbox.Attribute
			if drawY == v.getCursorY() {
				fg = termbox.ColorWhite
				bg = termbox.ColorGreen
			} else {
				fg = termbox.ColorDefault
				bg = termbox.ColorDefault
			}
			tbPrint(drawX, drawY, fg, bg, drawStr)
		}
	}
}
