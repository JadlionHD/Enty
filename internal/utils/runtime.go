package utils

import "runtime"

func (u *utils) GetUserOS() string {
	return runtime.GOOS
}

func (u *utils) GetUserArch() string {
	return runtime.GOARCH
}
