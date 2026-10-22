package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config struct {
	// add struct tags to specify the name of the key in the YAML file
	GlobalURL            string `yaml:"global_url" flag:"global-url" env:"GLOBAL_URL"`
	ArchiveDlCommandPath string `yaml:"archive_dl_command_path" flag:"archive-dl-command-path" env:"ARCHIVE_DL_COMMAND_PATH"`

	DatabaseName     string `yaml:"database_name" flag:"database-name" env:"DATABASE_NAME" default:"lbpsearch"`
	DatabaseUser     string `yaml:"database_user" flag:"database-user" env:"DATABASE_USER" default:"postgres"`
	DatabasePassword string `yaml:"database_password" flag:"database-password" env:"DATABASE_PASSWORD" default:""`
	DatabaseHost     string `yaml:"database_host" flag:"database-host" env:"DATABASE_HOST" default:"localhost"`
	DatabasePort     string `yaml:"database_port" flag:"database-port" env:"DATABASE_PORT" default:"5432"`
	DatabaseSSLMode  string `yaml:"database_ssl_mode" flag:"database-ssl-mode" env:"DATABASE_SSL_MODE" default:"disable"`

	ShowSponsorMessage bool `yaml:"show_sponsor_message" flag:"show-sponsor-message" env:"SHOW_SPONSOR_MESSAGE" default:"false"`

	HeaderInjection string `yaml:"header_injection" flag:"header-injection" env:"HEADER_INJECTION"`

	ArchiveBackend string `yaml:"storage_backend" flag:"storage-backend" env:"STORAGE_BACKEND"`
	CacheBackend   string `yaml:"cache_backend" flag:"cache-backend" env:"CACHE_BACKEND"`

	CachePath string `flag:"cache-path" yaml:"cache_path" env:"CACHE_PATH"`

	S3Endpoint    string `yaml:"s3_endpoint" flag:"s3-endpoint" env:"S3_ENDPOINT"`
	S3AccessKey   string `yaml:"s3_access_key" flag:"s3-access-key" env:"S3_ACCESS_KEY"`
	S3SecretKey   string `yaml:"s3_secret_key" flag:"s3-secret-key" env:"S3_SECRET_KEY"`
	ArchiveBucket string `yaml:"archive_bucket" flag:"archive-bucket" env:"ARCHIVE_BUCKET"`
	CacheBucket   string `yaml:"cache_bucket" flag:"cache-bucket" env:"CACHE_BUCKET"`
	S3Region      string `yaml:"s3_region" flag:"s3-region" env:"S3_REGION"`
}

// LoadConfig reads the configuration file and returns a Config struct.
func LoadConfig() (*Config, error) {

	cfg := new(Config)
	loader := aconfig.LoaderFor(cfg, aconfig.Config{
		EnvPrefix: "LBPSEARCH",
		Files:     []string{"./lbpsearch.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if err := loader.Load(); err != nil {
		return nil, err
	}

	return cfg, nil
}
