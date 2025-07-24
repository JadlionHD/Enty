package utils

import (
	"os"
	"path/filepath"
)

func (u *utils) GetTempDirectory() ([]string, error) {

	var dirPath = filepath.Join(u.GetPwd(), PATH_TEMP)
	exist := u.IsDirExist(dirPath)

	if !exist {
		u.Mkdir(dirPath)
	}

	var files []string

	err := filepath.Walk(PATH_TEMP, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (u *utils) IsDirExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (u *utils) Mkdir(dir string) error {
	return os.Mkdir(dir, os.ModePerm)
}
