package utils

import "context"

type utils struct {
	ctx context.Context
}

func Utils() *utils {
	return &utils{}
}

func (u *utils) Start(ctx context.Context) {
	u.ctx = ctx
}
