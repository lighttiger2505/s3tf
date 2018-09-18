package view

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	termbox "github.com/nsf/termbox-go"
)

type DetailView struct {
	Render
	Key   string
	Obj   *s3.GetObjectOutput
	Layer *Layer
}

func NewDetailView(x, y, width, height int) *DetailView {
	return &DetailView{
		Layer: NewLayer(x, y, width, height),
	}
}

func (v *DetailView) getContents() []string {
	base := `%v

    LastModified: %v
    Size: %v B
    ETag: %v
    Tags: %v`
	res := fmt.Sprintf(
		base,
		v.Key,
		aws.TimeValue(v.Obj.LastModified),
		aws.Int64Value(v.Obj.ContentLength),
		aws.StringValue(v.Obj.ETag),
		aws.Int64Value(v.Obj.TagCount),
	)
	return strings.Split(res, "\n")
}

func (v *DetailView) Up() int {
	return v.Layer.UpCursor(1)
}

func (v *DetailView) Down() int {
	lines := v.getContents()
	return v.Layer.DownCursor(1, len(lines))
}

func (v *DetailView) HalfPageUp() int {
	return v.Layer.HalfPageUpCursor()
}

func (v *DetailView) HalfPageDown() int {
	return v.Layer.HalfPageDownCursor(len(v.getContents()))
}

func (v *DetailView) Draw() {
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
