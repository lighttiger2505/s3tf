package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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
	items []*MenuItem
	layer *Layer
}

func (v *MenuView) Draw() {
	v.layer.DrawBackGround(termbox.ColorDefault, termbox.ColorDefault)

	lines := []string{}
	for _, item := range v.items {
		lines = append(lines, item.toString())
	}
	v.layer.DrawContents(
		lines,
		termbox.ColorWhite,
		termbox.ColorGreen,
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}

func (v *MenuView) getCursorItem() *MenuItem {
	return v.items[v.layer.cursorPos.Y]
}

func (v *MenuView) up() int {
	return v.layer.UpCursor(1)
}

func (v *MenuView) down() int {
	return v.layer.DownCursor(1, len(v.items))
}

type DetailView struct {
	Render
	key   string
	obj   *s3.GetObjectOutput
	layer *Layer
}

func (v *DetailView) getContents() []string {
	base := `%v

    LastModified: %v
    Size: %v B
    ETag: %v
    Tags: %v`
	res := fmt.Sprintf(
		base,
		v.key,
		aws.TimeValue(v.obj.LastModified),
		aws.Int64Value(v.obj.ContentLength),
		aws.StringValue(v.obj.ETag),
		aws.Int64Value(v.obj.TagCount),
	)
	return strings.Split(res, "\n")
}

func (v *DetailView) up() int {
	return v.layer.UpCursor(1)
}

func (v *DetailView) down() int {
	lines := v.getContents()
	return v.layer.DownCursor(1, len(lines))
}

func (v *DetailView) halfPageUp() int {
	return v.layer.HalfPageUpCursor()
}

func (v *DetailView) halfPageDown() int {
	return v.layer.HalfPageDownCursor(len(v.getContents()))
}

func (v *DetailView) Draw() {
	v.layer.DrawBackGround(termbox.ColorDefault, termbox.ColorDefault)

	lines := v.getContents()
	v.layer.DrawContents(
		lines,
		termbox.ColorWhite,
		termbox.ColorGreen,
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}

type DownloadObject struct {
	filename     string
	s3Path       string
	downloadPath string
}

type DownloadView struct {
	Render
	layer   *Layer
	objects []*DownloadObject
}

func (v *DownloadView) getContents() []string {
	drawLines := []string{}
	for _, object := range v.objects {
		tmpLine := strings.Join(
			[]string{
				object.filename,
				object.s3Path,
				object.downloadPath,
			},
			" ",
		)
		drawLines = append(drawLines, tmpLine)
	}
	return drawLines
}

func (v *DownloadView) up() int {
	return v.layer.UpCursor(1)
}

func (v *DownloadView) down() int {
	lines := v.getContents()
	return v.layer.DownCursor(1, len(lines))
}

func (v *DownloadView) halfPageUp() int {
	return v.layer.HalfPageUpCursor()
}

func (v *DownloadView) halfPageDown() int {
	return v.layer.HalfPageDownCursor(len(v.getContents()))
}

func (v *DownloadView) Draw() {
	v.layer.DrawBackGround(termbox.ColorDefault, termbox.ColorDefault)

	lines := v.getContents()
	v.layer.DrawContents(
		lines,
		termbox.ColorWhite,
		termbox.ColorGreen,
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}
