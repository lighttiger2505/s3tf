package main

import (
	"fmt"
	"log"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type Provider struct {
	EventHandler
	navigator      *Node
	bucket         string
	listView       *ListView
	navigationView *NavigationView
	statusView     *StatusView
}

func NewProvider() *Provider {
	p := &Provider{}
	p.Init()
	return p
}

func (p *Provider) Init() {
	// Init s3 data structure
	rootNode := NewNode("", nil, ListBuckets())
	width, height := termbox.Size()

	p.navigator = rootNode

	listView := &ListView{}
	listView.objects = p.navigator.objects
	listView.key = p.navigator.key
	listView.win = newWindow(0, 1, width, height-2)
	listView.cursorPos = newPosition(0, 0)
	listView.drawPos = newPosition(0, 0)
	p.listView = listView

	navigationView := &NavigationView{}
	navigationView.win = newWindow(0, 0, width, 1)
	p.navigationView = navigationView

	statusView := &StatusView{}
	statusView.win = newWindow(0, height-1, width, 1)
	p.statusView = statusView
}

func (p *Provider) Update() {
	p.navigationView.SetKey(p.listView.bucket, p.navigator.key)
}

func (p *Provider) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()
	p.listView.Draw()
	p.navigationView.Draw()
	p.statusView.Draw()
}

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

func (n *Node) IsBucketRoot() bool {
	if n.IsRoot() {
		return false
	}
	if n.parent.IsRoot() {
		return true
	}
	return false
}

func (n *Node) GetType() S3ListType {
	if n.IsRoot() {
		return BucketList
	}
	if n.IsBucketRoot() {
		return BucketRootList
	}
	return ObjectList
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

type S3ListType int

const (
	BucketList S3ListType = iota //0
	BucketRootList
	ObjectList
)

type ListView struct {
	Render
	// EventHandler
	// navigator *Node
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

func (w *Provider) Handle(ev termbox.Event) {
	if ev.Ch == 'j' {
		w.navigator.position = w.listView.down()
	} else if ev.Ch == 'k' {
		w.navigator.position = w.listView.up()
	} else if ev.Ch == 'h' {
		if !w.navigator.IsRoot() {
			w.loadPrev()
		}
	} else if ev.Ch == 'r' {
		w.reload()
	} else if ev.Ch == 'w' {
		w.download()
	} else if ev.Ch == 'l' || ev.Key == termbox.KeyEnter {
		obj := w.listView.getCursorObject()
		w.open(obj)
	}
}

func (w *ListView) up() int {
	if w.cursorPos.Y > 0 {
		w.cursorPos.Y--
		// w.navigator.position = w.cursorPos.Y
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
		// w.navigator.position = w.cursorPos.Y
	}
	if w.cursorPos.Y > (w.drawPos.Y + w.win.Box.Height - 1) {
		w.drawPos.Y = w.cursorPos.Y - w.win.Box.Height + 1
	}
	log.Printf("Down. CursorPosition:%d, DrawPosition:%d", w.cursorPos.Y, w.drawPos.Y)
	return w.cursorPos.Y
}

func (w *Provider) download() {
	obj := w.listView.getCursorObject()
	bucketName := w.bucket
	path := "s3://" + strings.Join([]string{bucketName, obj.Name}, "/")
	switch obj.ObjType {
	case Bucket:
		w.statusView.msg = fmt.Sprintf("%s is can't download. download command is file only", path)
	case Dir:
		w.statusView.msg = fmt.Sprintf("%s is can't download. download command is file only", path)
	case PreDir:
		w.statusView.msg = fmt.Sprintf("%s is can't download. download command is file only", path)
	case Object:
		DownloadObject(bucketName, obj.Name)
		path := "s3://" + strings.Join([]string{bucketName, obj.Name}, "/")
		w.statusView.msg = fmt.Sprintf("download complate. %s", path)
	default:
		log.Println("Invalid s3 object type")
	}
}

func (w *Provider) reload() {
	if w.navigator.IsRoot() {
		w.navigator.objects = ListBuckets()
		w.listView.objects = w.navigator.objects
		return
	}

	if w.navigator.IsBucketRoot() {
		w.navigator.objects = ListObjects(w.bucket, "")
		w.listView.objects = w.navigator.objects
		return
	}

	w.navigator.objects = ListObjects(w.bucket, w.navigator.key)
	w.listView.objects = w.navigator.objects
}

func (w *Provider) open(obj *S3Object) {
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

func (w *Provider) moveNext(key string) {
	child := w.navigator.GetChild(key)
	w.navigator = child
	w.listView.updateList(child)
	log.Printf("Move next. child:%s", child.key)
}

func (w *Provider) loadNext(key string, objects []*S3Object) {
	parent := w.navigator
	child := NewNode(key, parent, objects)
	parent.AddChild(key, child)
	w.navigator = child
	w.listView.updateList(child)
	log.Printf("Load next. parent:%s, child:%s", parent.key, child.key)
}

func (w *Provider) loadPrev() {
	parent := w.navigator.parent
	w.navigator = parent
	w.bucket = ""
	w.listView.updateList(parent)
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

type StatusView struct {
	Render
	msg string
	win *Window
}

func (w *StatusView) Draw() {
	str := PadRight(w.msg, w.win.Box.Width, " ")
	tbPrint(0, w.win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
