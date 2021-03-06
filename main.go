package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/lighttiger2505/s3tf/internal"
	"github.com/lighttiger2505/s3tf/model"
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

func isFileExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return err == nil || !os.IsNotExist(err)
}

func run(c *cli.Context) error {
	// Create and output log file
	logpath := filepath.Join(internal.GetXDGConfigPath(), "debug.log")
	logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile))

	// Set flags
	model.MockFlag = c.Bool("mock")

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	provider := NewProvider()
	provider.Loop()
	return nil
}
