package main

import (
	"os/exec"
	"runtime"
)

func Open(path string) error {
	opener := getOpener(runtime.GOOS)
	c := exec.Command(opener, path)
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func getOpener(goos string) (opener string) {
	switch goos {
	case "darwin":
		opener = "open"
	case "windows":
		opener = "cmd /c start"
	default:
		opener = "xdg-open"
	}
	return
}
