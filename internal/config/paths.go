package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JadlionHD/Enty/internal/configwatch"
)

var (
	livePathsConfig     *PathsConfigManager
	livePathsConfigLock sync.RWMutex
	watcherStopFunc     func()
)

// LivePathsConfigManager returns a singleton manager that auto-reloads on file changes.
func LivePathsConfigManager() *PathsConfigManager {
	livePathsConfigLock.RLock()
	if livePathsConfig != nil {
		defer livePathsConfigLock.RUnlock()
		return livePathsConfig
	}
	livePathsConfigLock.RUnlock()

	livePathsConfigLock.Lock()
	defer livePathsConfigLock.Unlock()
	if livePathsConfig == nil {
		livePathsConfig = NewPathsConfigManager(filepath.Join("config", "paths.json"))
		_ = livePathsConfig.LoadConfig()
		if watcherStopFunc != nil {
			watcherStopFunc()
		}
		watcherStopFunc = configwatch.WatchConfigFile(filepath.Join("config", "paths.json"), 2*time.Second, func() {
			livePathsConfigLock.Lock()
			_ = livePathsConfig.LoadConfig()
			livePathsConfigLock.Unlock()
		})
	}
	return livePathsConfig
}

// ServicePathsConfig holds the configuration for service-specific paths
type ServicePathsConfig struct {
	ServicePaths      map[string]string `json:"servicePaths"`
	DefaultPaths      []string          `json:"defaultPaths"`
	StandardUnixPaths []string          `json:"standardUnixPaths,omitempty"`
}

// PathsConfigManager manages the service paths configuration
type PathsConfigManager struct {
	config     *ServicePathsConfig
	configPath string
}

// NewPathsConfigManager creates a new paths configuration manager
func NewPathsConfigManager(configPath string) *PathsConfigManager {
	return &PathsConfigManager{
		configPath: configPath,
	}
}

// LoadConfig loads the paths configuration from the JSON file
func (pcm *PathsConfigManager) LoadConfig() error {
	if pcm.configPath == "" {
		// Use default config path
		pcm.configPath = filepath.Join("config", "paths.json")
	}

	// Check if file exists
	if _, err := os.Stat(pcm.configPath); os.IsNotExist(err) {
		return fmt.Errorf("paths configuration file not found: %s", pcm.configPath)
	}

	// Read the configuration file
	data, err := os.ReadFile(pcm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read paths config file: %w", err)
	}

	// Parse JSON
	var config ServicePathsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse paths config JSON: %w", err)
	}

	pcm.config = &config
	return nil
}

// GetServicePath returns the path for a specific service
func (pcm *PathsConfigManager) GetServicePath(serviceName string) (string, bool) {
	if pcm.config == nil {
		return "", false
	}

	path, exists := pcm.config.ServicePaths[strings.ToLower(serviceName)]
	return path, exists
}

// GetDefaultPaths returns the default system paths
func (pcm *PathsConfigManager) GetDefaultPaths() []string {
	if pcm.config == nil {
		return []string{}
	}
	return pcm.config.DefaultPaths
}

// GetStandardUnixPaths returns the standard Unix paths
func (pcm *PathsConfigManager) GetStandardUnixPaths() []string {
	if pcm.config == nil || len(pcm.config.StandardUnixPaths) == 0 {
		// Return default standard Unix paths if not configured
		return []string{"/usr/local/bin", "/usr/bin", "/bin"}
	}
	return pcm.config.StandardUnixPaths
}

// BuildIsolatedPath creates an isolated PATH environment variable for a specific service
func (pcm *PathsConfigManager) BuildIsolatedPath(serviceName string) string {
	// Ensure config is loaded
	if pcm.config == nil {
		_ = pcm.LoadConfig()
	}

	var pathComponents []string

	// Add service-specific path if it exists
	if servicePath, exists := pcm.GetServicePath(serviceName); exists {
		pathComponents = append(pathComponents, servicePath)
	}

	// Add default paths
	pathComponents = append(pathComponents, pcm.GetDefaultPaths()...)

	// Join with platform-specific separator
	return strings.Join(pathComponents, string(os.PathListSeparator))
}

// ValidateConfig validates the paths configuration
func (pcm *PathsConfigManager) ValidateConfig() []string {
	var warnings []string

	if pcm.config == nil {
		warnings = append(warnings, "Configuration not loaded")
		return warnings
	}

	// Check if service paths exist
	for serviceName, servicePath := range pcm.config.ServicePaths {
		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("Service path for '%s' does not exist: %s", serviceName, servicePath))
		}
	}

	// Check if default paths exist
	for _, defaultPath := range pcm.config.DefaultPaths {
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("Default path does not exist: %s", defaultPath))
		}
	}

	return warnings
}

// GetAllServices returns all configured service names
func (pcm *PathsConfigManager) GetAllServices() []string {
	if pcm.config == nil {
		return []string{}
	}

	services := make([]string, 0, len(pcm.config.ServicePaths))
	for serviceName := range pcm.config.ServicePaths {
		services = append(services, serviceName)
	}
	return services
}

// GetAllServicePaths returns the map of all service names to their paths
func (pcm *PathsConfigManager) GetAllServicePaths() map[string]string {
	if pcm.config == nil {
		return map[string]string{}
	}
	return pcm.config.ServicePaths
}
