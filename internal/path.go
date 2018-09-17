package internal

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func GetXDGConfigPath() string {
	home, _ := homedir.Dir()
	configdir := filepath.Join(home, ".config", "s3tf")
	if !IsFileExist(configdir) {
		os.Mkdir(configdir, os.FileMode(0755))
	}
	return configdir
}

func IsFileExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return err == nil || !os.IsNotExist(err)
}
