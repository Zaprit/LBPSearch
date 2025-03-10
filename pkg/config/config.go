package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	// add struct tags to specify the name of the key in the YAML file
	GlobalURL            string `yaml:"global_url"`
	ArchiveDlCommandPath string `yaml:"archive_dl_command_path"`

	DatabaseName     string `yaml:"database_name"`
	DatabaseUser     string `yaml:"database_user"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseHost     string `yaml:"database_host"`
	DatabasePort     string `yaml:"database_port"`
	DatabaseSSLMode  string `yaml:"database_ssl_mode"`

	ShowSponsorMessage bool `yaml:"show_sponsor_message"`

	HeaderInjection string `yaml:"header_injection"`

	ArchiveBackend string `yaml:"storage_backend"`
	CacheBackend   string `yaml:"cache_backend"`

	CachePath   string `yaml:"cache_path"`
	ArchivePath string `yaml:"archive_path"`

	S3Endpoint    string `yaml:"s3_endpoint"`
	S3AccessKey   string `yaml:"s3_access_key"`
	S3SecretKey   string `yaml:"s3_secret_key"`
	ArchiveBucket string `yaml:"archive_bucket"`
	CacheBucket   string `yaml:"cache_bucket"`
	S3Region      string `yaml:"s3_region"`
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
