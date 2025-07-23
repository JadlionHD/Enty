package config

import (
	"context"
	"log"
)

type configs struct {
	ctx context.Context
}

func Config() *configs {
	return &configs{}
}

func (c *configs) Start(ctx context.Context) {
	c.ctx = ctx

	// Initialize and validate paths configuration
	pathsConfig := LivePathsConfigManager()
	if warnings := pathsConfig.ValidateConfig(); len(warnings) > 0 {
		log.Printf("Paths configuration validation warnings:")
		for _, warning := range warnings {
			log.Printf("  - %s", warning)
		}
	}
	log.Printf("Paths configuration loaded successfully")
}
