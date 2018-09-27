package main

import (
	termbox "github.com/nsf/termbox-go"
)

func tbPrint(x, y int, fg, bg termbox.Attribute, msg string) {
	for i, c := range msg {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

func PadRight(str string, length int, padChar string) string {
	return str + times(padChar, length-len(str))
}
