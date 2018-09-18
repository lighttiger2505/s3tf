package view

import (
	"strings"

	"github.com/lighttiger2505/s3tf/model"
	termbox "github.com/nsf/termbox-go"
)

type DownloadView struct {
	Render
	Layer   *Layer
	Objects []*model.DownloadItem
}

func NewDownloadView(x, y, width, height int) *DownloadView {
	return &DownloadView{
		Layer: NewLayer(x, y, width, height),
	}
}

func (v *DownloadView) getContents() []string {
	drawLines := []string{}
	for _, object := range v.Objects {
		tmpLine := strings.Join(
			[]string{
				object.Filename,
				object.S3Path,
				object.DownloadPath,
			},
			" ",
		)
		drawLines = append(drawLines, tmpLine)
	}
	return drawLines
}

func (v *DownloadView) Up() int {
	return v.Layer.UpCursor(1)
}

func (v *DownloadView) Down() int {
	lines := v.getContents()
	return v.Layer.DownCursor(1, len(lines))
}

func (v *DownloadView) HalfPageUp() int {
	return v.Layer.HalfPageUpCursor()
}

func (v *DownloadView) HalfPageDown() int {
	return v.Layer.HalfPageDownCursor(len(v.getContents()))
}

func (v *DownloadView) Draw() {
	v.Layer.DrawBackGround(termbox.ColorDefault, termbox.ColorDefault)

	lines := v.getContents()
	v.Layer.DrawContents(
		lines,
		termbox.ColorWhite,
		termbox.ColorGreen,
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}
