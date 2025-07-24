package utils

import (
	"os"
	"runtime"
)

func (u *utils) GetUserOS() string {
	return runtime.GOOS
}

func (u *utils) GetUserArch() string {
	return runtime.GOARCH
}

func (u *utils) GetPwd() string {
	dir, err := os.Getwd()

	if err != nil {
		return ""
	}

	return dir
}
