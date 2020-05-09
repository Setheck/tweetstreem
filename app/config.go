package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigFormat = "json"
)

var (
	ConfigPath = ""
	ConfigFile = ".tweetstreem"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	if ConfigPath == "" {
		ConfigPath = home
	}
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigFormat)
	viper.AddConfigPath(ConfigPath)
}

func loadConfig(obj interface{}) {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("failed to read config file:", err)
		return
	}
	if err := viper.UnmarshalKey("config", &obj); err != nil {
		fmt.Println("unmarshalling config failed:", err)
	}
}

func saveConfig(obj interface{}) {
	viper.Set("config", obj)
	savePath := filepath.Join(ConfigPath, fmt.Sprint(ConfigFile, ".", ConfigFormat))
	if err := viper.WriteConfigAs(savePath); err != nil {
		log.Println("saving config to", savePath, "failed:", err)
	}
}
