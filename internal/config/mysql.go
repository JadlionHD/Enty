package config

import (
	"encoding/json"
	"log/slog"
	"os"
)

type ConfigVersionMySQL struct {
	Mysql []ConfigArchInfoMySQL `json:"mysql"`
}

type ConfigOSMySQL string

const (
	ConfigOSMySQLWindows ConfigOSMySQL = "Windows"
	ConfigOSMySQLLinux   ConfigOSMySQL = "Linux"
	ConfigOSMySQLMacOS   ConfigOSMySQL = "macOS"
)

type ConfigArchInfoMySQL struct {
	Os   ConfigOSMySQL     `json:"os"`
	Data []ConfigDataMySQL `json:"data"`
}

type ConfigDataMySQL struct {
	Version string  `json:"version"`
	Gpg     *string `json:"gpg,omitempty"`
	Link    string  `json:"link"`
}

func ReadConfig() (ConfigVersionMySQL, error) {
	var config ConfigVersionMySQL

	file, err := os.Open("config/mysql.json")
	if err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(-1)
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	jsonParser.Decode(&config)
	return config, nil
}
