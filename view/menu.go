package view

import (
	"fmt"

	termbox "github.com/nsf/termbox-go"
)

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
	Command   MenuCommand
}

func NewMenuItem(name, shorthand, detail string, command MenuCommand) *MenuItem {
	return &MenuItem{
		name:      name,
		shorthand: shorthand,
		detail:    detail,
		Command:   command,
	}
}

func (i *MenuItem) toString() string {
	return fmt.Sprintf("(%s)%s %s", i.shorthand, i.name, i.detail)
}

type MenuView struct {
	Render
	items []*MenuItem
	Layer *Layer
}

func NewMenuView(x, y, width, height int) *MenuView {
	view := &MenuView{
		Layer: NewLayer(x, y, width, height),
	}
	view.items = []*MenuItem{
		NewMenuItem("download", "w", "download file.", CommandDownload),
		NewMenuItem("open", "o", "open file.", CommandOpen),
		NewMenuItem("edit", "e", "open editor by file.", CommandEdit),
	}
	return view
}

func (v *MenuView) Draw() {
	v.Layer.DrawBackGround(termbox.ColorDefault, termbox.ColorDefault)

	lines := []string{}
	for _, item := range v.items {
		lines = append(lines, item.toString())
	}
	v.Layer.DrawContents(
		lines,
		termbox.ColorWhite,
		termbox.ColorGreen,
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}

func (v *MenuView) GetCursorItem() *MenuItem {
	return v.items[v.Layer.cursorPos.Y]
}

func (v *MenuView) Up() int {
	return v.Layer.UpCursor(1)
}

func (v *MenuView) Down() int {
	return v.Layer.DownCursor(1, len(v.items))
}
