package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	configOnce   sync.Once
	cachedConfig *viper.Viper
	log          *zap.Logger
)

func init() {
	log, _ = zap.NewProduction()
}

func GetConfig() *viper.Viper {
	configOnce.Do(func() {
		configPath := os.Getenv("APP_CONF")
		if configPath == "" {
			configPath = filepath.Join(getAppRootPath(), "config", "local.yml") // Default path relative to root
		}
		log.Info("Using config path", zap.String("path", configPath))
		cachedConfig = getConfig(configPath)
	})
	return cachedConfig
}

func getConfig(path string) *viper.Viper {
	conf := viper.New()
	conf.SetConfigFile(path)
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// Enable env var overrides — all YAML keys overridable via uppercased env vars
	// (dots replaced with underscores, e.g. HTTP_PORT → http.port)
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	conf.AutomaticEnv()

	// Conventional bindings — only applied if env var is non-empty
	if v := os.Getenv("DATABASE_URL"); v != "" {
		conf.Set("db.connectionString", v)
	}
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		conf.Set("messaging.connectionString", v)
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		conf.Set("security.jwt_secret", v)
	}

	return conf
}

func getAppRootPath() string {
	dir, err := os.Getwd() // Get current working directory
	if err != nil {
		panic(err)
	}
	return dir
}
