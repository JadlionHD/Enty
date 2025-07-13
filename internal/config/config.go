package config

import "context"

type configs struct {
	ctx context.Context
}

func Config() *configs {
	return &configs{}
}

func (c *configs) Start(ctx context.Context) {
	c.ctx = ctx
}
