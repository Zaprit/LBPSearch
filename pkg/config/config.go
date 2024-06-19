package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	// add struct tags to specify the name of the key in the YAML file
	CachePath            string `yaml:"cache_path"`
	GlobalURL            string `yaml:"global_url"`
	ArchivePath          string `yaml:"archive_path"`
	ArchiveDlCommandPath string `yaml:"archive_dl_command_path"`

	DatabaseName     string `yaml:"database_name"`
	DatabaseUser     string `yaml:"database_user"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseHost     string `yaml:"database_host"`
	DatabasePort     string `yaml:"database_port"`
	DatabaseSSLMode  string `yaml:"database_ssl_mode"`

	ShowSponsorMessage bool `yaml:"show_sponsor_message"`
}

// LoadConfig reads the configuration file and returns a Config struct.
func LoadConfig() (*Config, error) {
	bytes, err := os.ReadFile("lbpsearch.yaml")
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = yaml.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
