package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lighttiger2505/s3tf/model"
	"github.com/lighttiger2505/s3tf/view"
	termbox "github.com/nsf/termbox-go"
)

type eventAction string

const (
	// move cursor
	actQuit     = "quit"
	actUp       = "up"
	actDown     = "down"
	actHalfUp   = "half-up"
	actHalfDown = "half-down"
	// s3 control
	actReloadDir      = "reload-dir"
	actMoveNextDir    = "move-next-dir"
	actMovePrevDir    = "move-prev-dir"
	actDownloadObject = "download-object"
	actOpenObject     = "open-object"
	actEditObject     = "edit-object"
	// move view
	actOpenMenu     = "open-menu"
	actOpenDetail   = "open-detail"
	actOpenDownload = "open-download"
	// Menu view action
	actDoMenuAction = "do-menu-action"
)

var chMapOnList = map[rune]eventAction{
	'q': actQuit,
	'k': actUp,
	'j': actDown,
	'h': actMovePrevDir,
	'l': actMoveNextDir,
	'r': actReloadDir,
	'w': actDownloadObject,
	'o': actOpenObject,
	'e': actEditObject,
	'm': actOpenMenu,
	'n': actOpenDownload,
}
var keyMapOnList = map[termbox.Key]eventAction{
	termbox.KeyEsc:       actQuit,
	termbox.KeyArrowUp:   actUp,
	termbox.KeyCtrlP:     actUp,
	termbox.KeyArrowDown: actDown,
	termbox.KeyCtrlN:     actDown,
	termbox.KeyCtrlU:     actHalfUp,
	termbox.KeyCtrlD:     actHalfDown,
	termbox.KeyEnter:     actMoveNextDir,
}
var chMapOnMenu = map[rune]eventAction{
	'q': actQuit,
	'm': actQuit,
	'k': actUp,
	'j': actDown,
}
var keyMapOnMenu = map[termbox.Key]eventAction{
	termbox.KeyEsc:       actQuit,
	termbox.KeyArrowUp:   actUp,
	termbox.KeyCtrlP:     actUp,
	termbox.KeyArrowDown: actDown,
	termbox.KeyCtrlN:     actDown,
	termbox.KeyCtrlU:     actHalfUp,
	termbox.KeyCtrlD:     actHalfDown,
	termbox.KeyEnter:     actDoMenuAction,
}
var chMapOnDetail = map[rune]eventAction{
	'q': actQuit,
	'k': actUp,
	'j': actDown,
}
var keyMapOnDetail = map[termbox.Key]eventAction{
	termbox.KeyEsc:       actQuit,
	termbox.KeyArrowUp:   actUp,
	termbox.KeyCtrlP:     actUp,
	termbox.KeyArrowDown: actDown,
	termbox.KeyCtrlN:     actDown,
	termbox.KeyCtrlU:     actHalfUp,
	termbox.KeyCtrlD:     actHalfDown,
}
var chMapOnDownload = map[rune]eventAction{
	'q': actQuit,
	'k': actUp,
	'j': actDown,
}
var keyMapOnDownload = map[termbox.Key]eventAction{
	termbox.KeyEsc:       actQuit,
	termbox.KeyArrowUp:   actUp,
	termbox.KeyCtrlP:     actUp,
	termbox.KeyArrowDown: actDown,
	termbox.KeyCtrlN:     actDown,
	termbox.KeyCtrlU:     actHalfUp,
	termbox.KeyCtrlD:     actHalfDown,
}

func getEventAction(
	ev termbox.Event,
	chMap map[rune]eventAction,
	keyMap map[termbox.Key]eventAction,
) eventAction {
	var res eventAction
	if val, ok := chMap[ev.Ch]; ok {
		res = val
	}
	if val, ok := keyMap[ev.Key]; ok {
		res = val
	}
	return res
}

type EventHandler interface {
	Handle(termbox.Event)
}

type ProviderStatus int

const (
	StateList ProviderStatus = iota //0
	StateMenu
	StateDetail
	StateDownload
)

type Provider struct {
	EventHandler
	status         ProviderStatus
	node           *model.Node
	bucket         string
	dllFile        *model.DownloadListFile
	listView       *ListView
	navigationView *NavigationView
	statusView     *view.StatusView
	menuView       *view.MenuView
	detailView     *view.DetailView
	downloadView   *view.DownloadView
}

func NewProvider() *Provider {
	p := &Provider{}
	p.Init()
	p.Update()
	p.Draw()
	return p
}

func (p *Provider) Init() {
	// Init s3 data structure
	rootNode := model.NewNode("", nil, model.ListBuckets())
	width, height := termbox.Size()
	halfWidth := width / 2
	halfHeight := height / 2

	p.status = StateList
	p.node = rootNode
	dllFile, err := model.LoadDownloadFile()
	if err != nil {
		panic("failed load download list file.")
	}
	p.dllFile = dllFile

	listView := &ListView{}
	listView.objects = p.node.Objects
	listView.key = p.node.Key
	listView.layer = NewLayer(0, 1, width, height-2)
	p.listView = listView

	navigationView := &NavigationView{}
	navigationView.win = newWindow(0, 0, width, 1)
	p.navigationView = navigationView

	p.statusView = view.NewStatusView(0, height-1, width, 1)
	p.menuView = view.NewMenuView(0, halfHeight, width, height-halfHeight)
	p.detailView = view.NewDetailView(halfWidth, 1, width-halfWidth, height-2)
	p.downloadView = view.NewDownloadView(0, 1, width, height-2)
}

func (p *Provider) Loop() {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			p.Handle(ev)
			p.Update()
		case termbox.EventError:
			panic(ev.Err)
		case termbox.EventInterrupt:
			return
		}
		p.Draw()
	}
}

func (p *Provider) Update() {
	p.navigationView.SetCurrentPath(p.bucket, p.node)
}

func (p *Provider) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()
	p.listView.Draw()
	p.navigationView.Draw()
	if p.status == StateMenu {
		p.menuView.Draw()
	}
	if p.status == StateDetail {
		p.detailView.Draw()
	}
	if p.status == StateDownload {
		p.downloadView.Draw()
	}
	p.statusView.Draw()
}

func (p *Provider) reload() {
	if p.node.IsRoot() {
		p.node.Objects = model.ListBuckets()
		p.listView.objects = p.node.Objects
		return
	}

	if p.node.IsBucketRoot() {
		p.node.Objects = model.ListObjects(p.bucket, "")
		p.listView.objects = p.node.Objects
		return
	}

	p.node.Objects = model.ListObjects(p.bucket, p.node.Key)
	p.listView.objects = p.node.Objects
}

func (p *Provider) download() {
	obj := p.listView.getCursorObject()
	bucketName := p.bucket
	switch obj.ObjType {
	case model.Object:
		filename := model.Filename(obj.Name)
		currentDir, _ := os.Getwd()
		downloadPath := filepath.Join(currentDir, filename)
		f, err := os.Create(downloadPath)
		if err != nil {
			log.Fatalf("failed create donwload reader, %v", err)
		}
		defer f.Close()

		model.Download(bucketName, obj.Name, f)
		s3Path := "s3://" + strings.Join([]string{bucketName, obj.Name}, "/")

		p.dllFile.Items = append(
			p.dllFile.Items,
			model.NewDownloadItem(
				filename,
				downloadPath,
				s3Path,
			),
		)
		if err := model.SaveDownloadFile(p.dllFile); err != nil {
			log.Fatalf("failed save download list file, %v", err)
		}

		p.statusView.Msg = fmt.Sprintf("download complate. %s", s3Path)
	default:
		log.Println("Invalid s3 object type")
	}
}

func (p *Provider) open() {
	obj := p.listView.getCursorObject()
	bucketName := p.bucket
	switch obj.ObjType {
	case model.Object:
		tempDir, _ := ioutil.TempDir("", "")
		f, err := os.Create(filepath.Join(tempDir, model.Filename(obj.Name)))
		if err != nil {
			log.Fatalf("failed create donwload reader, %v", err)
		}
		defer f.Close()

		model.Download(bucketName, obj.Name, f)
		if err := Open(f.Name()); err != nil {
			log.Fatalf("failed open file, %v", err)
		}

		path := "s3://" + strings.Join([]string{bucketName, obj.Name}, "/")
		p.statusView.Msg = fmt.Sprintf("open. %s", path)
	default:
		log.Println("Invalid s3 object type")
	}
}

func (p *Provider) edit() {
	obj := p.listView.getCursorObject()
	bucketName := p.bucket
	switch obj.ObjType {
	case model.Object:
		// download edit file on temporary file
		tempDir, _ := ioutil.TempDir("", "")
		f, err := os.Create(filepath.Join(tempDir, model.Filename(obj.Name)))
		if err != nil {
			log.Fatalf("failed create donwload reader, %v", err)
		}
		model.Download(bucketName, obj.Name, f)
		editFilePath := f.Name()
		f.Close()

		// termbox close and restert for edit
		termbox.Close()
		defer termbox.Init()
		OpenEditor(editFilePath)

		// update edited object
		editedf, err := os.Open(editFilePath)
		if err != nil {
			log.Fatalf("failed open edited file, %v", err)
		}
		model.Update(bucketName, obj.Name, editedf)

		path := "s3://" + strings.Join([]string{bucketName, obj.Name}, "/")
		p.statusView.Msg = fmt.Sprintf("edit. %s", path)
	default:
		log.Println("Invalid s3 object type")
	}
}

func (p *Provider) show(obj *model.S3Object) {
	switch obj.ObjType {
	case model.Bucket:
		bucketName := obj.Name
		p.bucket = bucketName
		if p.node.IsExistChildren(bucketName) {
			p.moveNext(bucketName)
			return
		}
		objects := model.ListObjects(bucketName, "")
		p.loadNext(bucketName, objects)
	case model.Dir:
		bucketName := p.bucket
		objectKey := obj.Name
		if p.node.IsExistChildren(objectKey) {
			p.moveNext(objectKey)
			return
		}
		objects := model.ListObjects(bucketName, objectKey)
		p.loadNext(objectKey, objects)
	case model.PreDir:
		p.loadPrev()
	case model.Object:
	default:
		log.Fatalln("Invalid s3 object type")
	}
}

func (p *Provider) moveNext(key string) {
	child := p.node.GetChild(key)
	p.node = child
	p.listView.updateList(child)
	log.Printf("Move next. child:%s", child.Key)
}

func (p *Provider) loadNext(key string, objects []*model.S3Object) {
	parent := p.node
	child := model.NewNode(key, parent, objects)
	parent.AddChild(key, child)
	p.node = child
	p.listView.updateList(child)
	log.Printf("Load next. parent:%s, child:%s", parent.Key, child.Key)
}

func (p *Provider) loadPrev() {
	parent := p.node.Parent
	p.node = parent
	p.listView.updateList(parent)
	log.Printf("Load prev. parent:%s", parent.Key)
}

func (p *Provider) menu() {
	p.status = StateMenu
}

func (p *Provider) detail(obj *model.S3Object) {
	p.status = StateDetail
	p.detailView.Obj = model.Detail(p.bucket, obj.Name)
	p.detailView.Key = obj.Name
}

func (p *Provider) openDownload() {
	p.status = StateDownload
	p.downloadView.Objects = p.dllFile.Items
}

func (p *Provider) Handle(ev termbox.Event) {
	switch p.status {
	case StateList:
		p.listEvent(ev)
	case StateMenu:
		p.menuEvent(ev)
	case StateDetail:
		p.detailEvent(ev)
	case StateDownload:
		p.downloadEvent(ev)
	}
}

func (p *Provider) listEvent(ev termbox.Event) {
	ea := getEventAction(ev, chMapOnList, keyMapOnList)
	if ea == "" {
		p.statusView.Msg = "no mapping key"
		return
	}

	switch ea {
	case actQuit:
		go func() {
			termbox.Interrupt()
			time.Sleep(1 * time.Second)
			panic("this should never run")
		}()
	case actDown:
		p.node.Position = p.listView.down()
	case actUp:
		p.node.Position = p.listView.up()
	case actHalfUp:
		p.listView.halfPageUp()
	case actHalfDown:
		p.listView.halfPageDown()
	case actOpenMenu:
		p.menu()
	case actOpenDetail:
		obj := p.listView.getCursorObject()
		p.detail(obj)
	case actOpenDownload:
		p.openDownload()
	case actMovePrevDir:
		if !p.node.IsRoot() {
			p.loadPrev()
		}
	case actMoveNextDir:
		obj := p.listView.getCursorObject()
		p.show(obj)
	case actReloadDir:
		p.reload()
	case actOpenObject:
		p.open()
	case actDownloadObject:
		p.download()
	case actEditObject:
		p.edit()
	default:
	}
}

func (p *Provider) menuEvent(ev termbox.Event) {
	ea := getEventAction(ev, chMapOnMenu, keyMapOnMenu)
	if ea == "" {
		p.statusView.Msg = "no mapping key"
		return
	}

	switch ea {
	case actQuit:
		p.status = StateList
	case actUp:
		p.menuView.Up()
	case actDown:
		p.menuView.Down()
	case actDoMenuAction:
		item := p.menuView.GetCursorItem()
		switch item.Command {
		case view.CommandDownload:
			p.download()
		case view.CommandOpen:
			p.open()
		case view.CommandEdit:
			p.edit()
		}
		p.status = StateList
	default:
	}
}

func (p *Provider) detailEvent(ev termbox.Event) {
	ea := getEventAction(ev, chMapOnDetail, keyMapOnDetail)
	if ea == "" {
		p.statusView.Msg = "no mapping key"
		return
	}

	switch ea {
	case actQuit:
		p.status = StateList
	case actUp:
		p.detailView.Up()
	case actDown:
		p.detailView.Down()
	case actHalfUp:
		p.detailView.HalfPageUp()
	case actHalfDown:
		p.detailView.HalfPageDown()
	default:
	}
}

func (p *Provider) downloadEvent(ev termbox.Event) {
	ea := getEventAction(ev, chMapOnDownload, keyMapOnDownload)
	if ea == "" {
		p.statusView.Msg = "no mapping key"
		return
	}

	switch ea {
	case actQuit:
		p.status = StateList
	case actUp:
		p.downloadView.Up()
	case actDown:
		p.downloadView.Down()
	case actHalfUp:
		p.downloadView.HalfPageUp()
	case actHalfDown:
		p.downloadView.HalfPageDown()
	default:
	}
}
