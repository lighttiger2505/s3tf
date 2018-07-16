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
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "mock, m",
			Usage: "S3 api request to mock server on localhost(minio)",
		},
	}
	app.Action = run
	return app
}

var (
	mockFlag bool
)

func run(c *cli.Context) error {
	// Init logging setting
	logfile, err := os.OpenFile("./debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile))

	// Set flags
	mockFlag = c.Bool("mock")

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	provider := NewProvider()
	provider.Draw()
mainloop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Ch == 'q' {
				break mainloop
			}
			provider.Update(ev)

		case termbox.EventError:
			panic(ev.Err)

		case termbox.EventInterrupt:
			break mainloop
		}
		provider.Draw()
	}
	return nil
}
