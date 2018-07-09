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

var bucketList *BucketList

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
	bucketList.Draw()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Ch == 'q' {
				break mainloop
			}
			bucketList.Handle(ev)

		case termbox.EventError:
			panic(ev.Err)

		case termbox.EventInterrupt:
			break mainloop
		}

		bucketList.Draw()
	}
	return nil
}

func Init() {
	bucketList = &BucketList{}
	bucketList.buckets = ListBuckets()
	width, height := termbox.Size()
	fmt.Println(width, height)
	bucketList.win = newWindow(0, 0, width, height)
	bucketList.cursorPos = newPosition(0, 0)
	bucketList.drawPos = newPosition(0, 0)
}
