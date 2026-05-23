package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func GetConfig() *viper.Viper {
	configPath := os.Getenv("APP_CONF")
	if configPath == "" {
		configPath = filepath.Join(getAppRootPath(), "config", "local.yml") // Default path relative to root
	}
	fmt.Println("Using config path:", configPath)
	return getConfig(configPath)
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
	if v := os.Getenv("HTTP_PORT"); v != "" {
		conf.Set("http.port", v)
	}
	if v := os.Getenv("HTTP_SECURE"); v != "" {
		conf.Set("http.secure", v)
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		conf.Set("grpc.port", v)
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		conf.Set("db.connectionString", v)
	}
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		conf.Set("messaging.connectionString", v)
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		conf.Set("security.jwt_secret", v)
	}
	if v := os.Getenv("JWT_ISSUER"); v != "" {
		conf.Set("security.jwt_issuer", v)
	}
	if v := os.Getenv("REFRESH_TOKEN_SECRET"); v != "" {
		conf.Set("security.refresh_secret", v)
	}
	if v := os.Getenv("GOOGLE_CLIENT_ID"); v != "" {
		conf.Set("google.client_id", v)
	}
	if v := os.Getenv("GOOGLE_CLIENT_SECRET"); v != "" {
		conf.Set("google.client_secret", v)
	}
	if v := os.Getenv("GOOGLE_REDIRECT_URL"); v != "" {
		conf.Set("google.redirect_url", v)
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
