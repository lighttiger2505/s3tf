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

type Node struct {
	key      string
	parent   *Node
	children map[string]*Node
	objects  []*S3Object
	position int
}

func NewNode(key string, parent *Node, objects []*S3Object) *Node {
	node := &Node{
		key:      key,
		parent:   parent,
		objects:  objects,
		children: map[string]*Node{},
	}
	if len(objects) > 1 {
		node.position = 1
	}
	return node
}

func (n *Node) IsRoot() bool {
	if n.parent == nil {
		return true
	}
	return false
}

func (n *Node) IsExistChildren(key string) bool {
	_, ok := n.children[key]
	return ok
}

func (n *Node) GetChild(key string) *Node {
	return n.children[key]
}

func (n *Node) AddChild(key string, node *Node) {
	n.children[key] = node
}

type ListView struct {
	Render
	EventHandler
	navigator *Node
	bucket    string
	win       *Window
	cursorPos *Position
	drawPos   *Position
}

func (w *ListView) Draw() {
	for i, obj := range w.navigator.objects {
		drawStr := obj.Name
		if w.navigator.parent == nil || !w.navigator.parent.IsRoot() {
			drawStr = strings.TrimPrefix(obj.Name, w.navigator.key)
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

func (w *ListView) Handle(ev termbox.Event) {
	if ev.Ch == 'j' {
		w.down()
	} else if ev.Ch == 'k' {
		w.up()
	} else if ev.Ch == 'h' {
		if !w.navigator.IsRoot() {
			w.loadPrev()
		}
	} else if ev.Ch == 'l' || ev.Key == termbox.KeyEnter {
		obj := w.navigator.objects[w.cursorPos.Y]
		w.open(obj)
	}
}

func (w *ListView) up() {
	if w.cursorPos.Y > 0 {
		w.cursorPos.Y--
		w.navigator.position = w.cursorPos.Y
	}
	if w.cursorPos.Y < w.drawPos.Y {
		w.drawPos.Y = w.cursorPos.Y
	}
	log.Printf("Up. CursorPosition:%d, DrawPosition:%d", w.cursorPos.Y, w.drawPos.Y)
}

func (w *ListView) down() {
	if w.cursorPos.Y < (len(w.navigator.objects) - 1) {
		w.cursorPos.Y++
		w.navigator.position = w.cursorPos.Y
	}
	if w.cursorPos.Y > (w.drawPos.Y + w.win.Box.Height - 1) {
		w.drawPos.Y = w.cursorPos.Y - w.win.Box.Height + 1
	}
	log.Printf("Down. CursorPosition:%d, DrawPosition:%d", w.cursorPos.Y, w.drawPos.Y)
}

func (w *ListView) open(obj *S3Object) {
	switch obj.ObjType {
	case Bucket:
		bucketName := obj.Name
		w.bucket = bucketName
		if w.navigator.IsExistChildren(bucketName) {
			w.moveNext(bucketName)
			return
		}
		objects := ListObjects(bucketName, "")
		w.loadNext(bucketName, objects)
	case Dir:
		bucketName := w.bucket
		objectKey := obj.Name
		if w.navigator.IsExistChildren(objectKey) {
			w.moveNext(objectKey)
			return
		}
		objects := ListObjects(bucketName, objectKey)
		w.loadNext(objectKey, objects)
	case PreDir:
		w.loadPrev()
	case Object:
	default:
		log.Fatalln("Invalid s3 object type")
	}
}

func (w *ListView) moveNext(key string) {
	child := w.navigator.GetChild(key)
	w.navigator = child
	w.cursorPos.Y = child.position
	log.Printf("Move next. child:%s", child.key)
}

func (w *ListView) loadNext(key string, objects []*S3Object) {
	parent := w.navigator
	child := NewNode(key, parent, objects)
	parent.AddChild(key, child)
	w.navigator = child
	w.cursorPos.Y = child.position
	log.Printf("Load next. parent:%s, child:%s", parent.key, child.key)
}

func (w *ListView) loadPrev() {
	parent := w.navigator.parent
	w.navigator = parent
	w.cursorPos.Y = parent.position
	w.bucket = ""
	log.Printf("Load prev. parent:%s", parent.key)
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
