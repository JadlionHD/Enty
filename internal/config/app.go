package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type ConfigVersionApp struct {
	App []ConfigArchInfoApp `json:"app"`
}

type ConfigOSApp string

const (
	ConfigOSAppWindows ConfigOSApp = "Windows"
	ConfigOSAppLinux   ConfigOSApp = "Linux"
	ConfigOSAppMacOS   ConfigOSApp = "macOS"
)

type ConfigArchInfoApp struct {
	Os   ConfigOSApp     `json:"os"`
	Data []ConfigDataApp `json:"data"`
}

type ConfigDataApp struct {
	Version string  `json:"version"`
	Gpg     *string `json:"gpg,omitempty"`
	Link    string  `json:"link"`
}

func (c *configs) GetAppConfig(configFile string) (*ConfigVersionApp, error) {
	var config ConfigVersionApp

	file, err := os.Open(filepath.Join("config", configFile))
	if err != nil {
		slog.Error("failed to read config", "error", err)
		return nil, err
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	jsonParser.Decode(&config)
	return &config, nil
}
