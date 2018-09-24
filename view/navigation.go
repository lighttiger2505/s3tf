package view

import (
	"fmt"
	"strings"

	"github.com/lighttiger2505/s3tf/model"
	termbox "github.com/nsf/termbox-go"
)

type NavigationView struct {
	Render
	currentPath string
	Win         *Window
}

func NewNavigationView(x, y, width, height int) *NavigationView {
	return &NavigationView{
		Win: newWindow(x, y, width, height),
	}
}

func (v *NavigationView) SetCurrentPath(bucket string, node *model.Node) {
	if node.IsRoot() {
		v.currentPath = "list bucket"
		return
	}

	showBucketName := fmt.Sprintf("s3://%s", bucket)
	if node.IsBucketRoot() {
		v.currentPath = showBucketName
	} else {
		v.currentPath = strings.Join([]string{showBucketName, node.Key}, "/")
	}
}

func (v *NavigationView) Draw() {
	str := PadRight(v.currentPath, v.Win.Box.Width, " ")
	tbPrint(0, v.Win.DrawY(0), termbox.ColorWhite, termbox.ColorBlue, str)
}
