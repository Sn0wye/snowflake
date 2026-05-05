package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	return conf
}

func getAppRootPath() string {
	dir, err := os.Getwd() // Get current working directory
	if err != nil {
		panic(err)
	}
	return dir
}
