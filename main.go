package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nsf/termbox-go"
)

var bucketList *BucketList

func Init() {
	bucketList = &BucketList{}
	bucketList.buckets = ListBuckets()
	width, height := termbox.Size()
	fmt.Println(width, height)
	bucketList.win = newWindow(0, 0, width, height)
	bucketList.cursorPos = newPosition(0, 0)
	bucketList.drawPos = newPosition(0, 0)

}

func main() {
	logfile, err := os.OpenFile("./debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile))

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

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
	termbox.Close()
}
