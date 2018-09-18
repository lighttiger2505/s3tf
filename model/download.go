package model

import (
	// 	"gopkg.in/yaml.v2"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/lighttiger2505/s3tf/internal"
	yaml "gopkg.in/yaml.v2"
)

type DownloadItem struct {
	Filename     string `yaml:"filename"`
	S3Path       string `yaml:"s3_path"`
	DownloadPath string `yaml:"download_path"`
}

func NewDownloadItem(filename, s3Path, downloadPath string) *DownloadItem {
	return &DownloadItem{
		Filename:     filename,
		S3Path:       s3Path,
		DownloadPath: downloadPath,
	}
}

type DownloadListFile struct {
	Items []*DownloadItem `yaml:"items"`
}

func GetDownloadFilePath() (string, error) {
	xdgConfigPath := internal.GetXDGConfigPath()
	dllPath := filepath.Join(xdgConfigPath, "downloads.yml")
	if !internal.IsFileExist(dllPath) {
		if err := createDownloadFile(dllPath); err != nil {
			return "", err
		}
	}
	return dllPath, nil
}

func LoadDownloadFile() (*DownloadListFile, error) {
	fpath, err := GetDownloadFilePath()
	if err != nil {
		return nil, err
	}

	file, err := open(fpath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); err != nil {
			err = cerr
		}
	}()

	res, err := read(file)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func SaveDownloadFile(dllFile *DownloadListFile) error {
	fpath, err := GetDownloadFilePath()
	if err != nil {
		return err
	}

	file, err := open(fpath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); err != nil {
			err = cerr
		}
	}()

	if err = write(file, dllFile); err != nil {
		return err
	}

	return nil
}

func createDownloadFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Failed create config file: %s", err.Error())
	}
	defer file.Close()

	config := &DownloadListFile{}
	out, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("Failed marshal config: %v", err.Error())
	}

	_, err = file.Write(out)
	if err != nil {
		return fmt.Errorf("Failed write config file: %s", err.Error())
	}

	return nil
}

func open(fpath string, flag int, perm os.FileMode) (*os.File, error) {
	if !internal.IsFileExist(fpath) {
		return nil, fmt.Errorf("Not exist config. path %s", fpath)
	}

	file, err := os.OpenFile(fpath, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("Filed open file. Error: %s", err.Error())
	}
	return file, nil
}

func read(r io.Reader) (*DownloadListFile, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed unmarshal yaml. Error: %s", err.Error())
	}

	dllFile := &DownloadListFile{}
	if err := yaml.Unmarshal(b, dllFile); err != nil {
		return nil, fmt.Errorf("Failed unmarshal yaml. \nError: %s \nBuffer: %s", err.Error(), string(b))
	}
	return dllFile, nil
}

func write(writer io.Writer, dllFile *DownloadListFile) error {
	out, err := yaml.Marshal(dllFile)
	if err != nil {
		return fmt.Errorf("Failed marshal config. Error: %v", err.Error())
	}

	if _, err = io.WriteString(writer, string(out)); err != nil {
		return fmt.Errorf("Failed write config file. Error: %s", err.Error())
	}
	return nil
}
