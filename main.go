package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nsf/termbox-go"
	"github.com/urfave/cli"
)

const (
	ExitCodeOK    int = iota //0
	ExitCodeError int = iota //1
)

var (
	listView       *ListView
	navigationView *NavigationView
)

func main() {
	err := newApp().Run(os.Args)
	var exitCode = ExitCodeOK
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		exitCode = ExitCodeError
	}
	os.Exit(exitCode)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "s3 terminal finder"
	app.HelpName = "s3tf"
	app.Usage = "AWS S3 TUI file manager"
	app.UsageText = "s3tf [options] <args>"
	app.Version = "0.0.1"
	app.Author = "lighttiger2505"
	app.Email = "lighttiger2505@gmail.com"
	app.Flags = []cli.Flag{}
	app.Action = run
	return app
}

func run(c *cli.Context) error {
	// Init logging setting
	logfile, err := os.OpenFile("./debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile))

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	Init()
	navigationView.SetKey(listView.bucket, listView.navigator.key)
	draw()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Ch == 'q' {
				break mainloop
			}
			listView.Handle(ev)
			navigationView.SetKey(listView.bucket, listView.navigator.key)

		case termbox.EventError:
			panic(ev.Err)

		case termbox.EventInterrupt:
			break mainloop
		}
		draw()
	}
	return nil
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()
	listView.Draw()
	navigationView.Draw()
}

func Init() {
	// Init s3 data structure
	rootNode := NewNode("", nil, ListBuckets())

	width, height := termbox.Size()

	listView = &ListView{}
	listView.navigator = rootNode
	listView.win = newWindow(0, 1, width, height-1)
	listView.cursorPos = newPosition(0, 0)
	listView.drawPos = newPosition(0, 0)

	navigationView = &NavigationView{}
	navigationView.win = newWindow(0, 0, width, 1)
}
