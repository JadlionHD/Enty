package utils

import (
	"context"

	"github.com/JadlionHD/Enty/internal/config"
)

type utils struct {
	ctx context.Context
}

func Utils() *utils {
	return &utils{}
}

func (u *utils) Start(ctx context.Context) {
	u.ctx = ctx
}

// GetAvailableServices returns a list of available services for terminal execution
func (u *utils) GetAvailableServices() []string {
	pathsConfig := config.LivePathsConfigManager()
	if pathsConfig == nil {
		return []string{}
	}
	return pathsConfig.GetAllServices()
}

// ValidateServicePath checks if a service path exists and is valid
func (u *utils) ValidateServicePath(serviceName string) (bool, string) {
	pathsConfig := config.LivePathsConfigManager()
	if pathsConfig == nil {
		return false, "Paths configuration not loaded"
	}

	servicePath, exists := pathsConfig.GetServicePath(serviceName)
	if !exists {
		return false, "Service not found in configuration"
	}

	warnings := pathsConfig.ValidateConfig()
	for _, warning := range warnings {
		if warning != "" && len(warning) > len(serviceName) && warning[:len(serviceName)] == serviceName {
			return false, warning
		}
	}

	return true, servicePath
}
