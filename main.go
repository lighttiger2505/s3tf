package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func draw(count, cursorPosition int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	for i := 0; i < count; i++ {
		s := fmt.Sprintf("count = %d", i)
		tbPrint(0, i, termbox.ColorDefault, termbox.ColorDefault, s)
	}
	termbox.SetCursor(0, cursorPosition)
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc)

	count := 10
	cursorPosition := 5
	draw(count, cursorPosition)
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch == '+' {
				count++
			} else if ev.Ch == 'j' {
				cursorPosition++
			} else if ev.Ch == 'k' {
				cursorPosition--
			} else if ev.Ch == '-' {
				count--
			} else if ev.Key == termbox.KeyEsc || ev.Ch == 'q' {
				break mainloop
			}

		case termbox.EventError:
			panic(ev.Err)

		case termbox.EventInterrupt:
			break mainloop
		}

		draw(count, cursorPosition)
	}
	termbox.Close()

	fmt.Println("Finished")
}
