package view

import (
	"strings"

	"github.com/lighttiger2505/s3tf/model"
	termbox "github.com/nsf/termbox-go"
)

type ListView struct {
	Render
	Key      string
	listType model.S3ListType
	Objects  []*model.S3Object
	Layer    *Layer
}

func NewListView(x, y, width, height int) *ListView {
	return &ListView{
		Layer: NewLayer(x, y, width, height),
	}
}

func (v *ListView) Draw() {
	for i, obj := range v.Objects {
		drawStr := obj.Name
		if v.listType == model.ObjectList {
			drawStr = strings.TrimPrefix(obj.Name, v.Key)
		}

		if i >= v.Layer.drawPos.Y {
			drawY := v.Layer.getDrawY(i)
			var fg, bg termbox.Attribute
			if drawY == v.Layer.getCursorY() {
				drawStr = PadRight(drawStr, v.Layer.win.Box.Width, " ")
				fg = termbox.ColorWhite
				bg = termbox.ColorGreen
			} else if model.Bucket == obj.ObjType || model.PreDir == obj.ObjType || model.Dir == obj.ObjType {
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

func (v *ListView) GetCursorObject() *model.S3Object {
	return v.Objects[v.Layer.cursorPos.Y]
}

func (v *ListView) UpdateList(node *model.Node) {
	v.Layer.cursorPos.Y = node.Position
	v.Objects = node.Objects
	v.Key = node.Key
	v.listType = node.GetType()
}

func (v *ListView) Up() int {
	return v.Layer.UpCursor(1)
}

func (v *ListView) Down() int {
	return v.Layer.DownCursor(1, len(v.Objects))
}

func (v *ListView) HalfPageUp() int {
	return v.Layer.HalfPageUpCursor()
}

func (v *ListView) HalfPageDown() int {
	return v.Layer.HalfPageDownCursor(len(v.Objects))
}
