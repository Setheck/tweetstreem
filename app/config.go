package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigFile   = ".tweetstreem"
	ConfigFormat = "json"
)

func init() {
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigFormat)
	viper.AddConfigPath("$HOME/") // TODO:(smt) how does this work on windows.
	viper.AddConfigPath(".")
}

func loadConfig(t interface{}) {
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		return
	}

	err = viper.UnmarshalKey("config", &t)
	if err != nil {
		fmt.Println(err)
	}
}

func saveConfig(c interface{}) {
	viper.Set("config", c)
	hd, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return
	}
	fileName := fmt.Sprint(ConfigFile, ".", ConfigFormat)
	if err := viper.WriteConfigAs(filepath.Join(hd, fileName)); err != nil {
		log.Println(err)
	}
}
