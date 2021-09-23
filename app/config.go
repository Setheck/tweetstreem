package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Setheck/tweetstreem/util"
	"github.com/spf13/viper"
)

const (
	configFormat = "json"
)

var (
	configPath = ""
	configFile = ".tweetstreem"
)

type Viper interface {
	SetConfigName(string)
	SetConfigType(string)
	AddConfigPath(string)
	ReadInConfig() error
	UnmarshalKey(string, interface{}, ...viper.DecoderConfigOption) error
	WriteConfigAs(string) error
	Set(string, interface{})
}

var tsViper Viper = viper.New()

func init() {
	home := util.MustString(os.UserHomeDir())
	if configPath == "" {
		configPath = home
	}
	tsViper.SetConfigName(configFile)
	tsViper.SetConfigType(configFormat)
	tsViper.AddConfigPath(configPath)
}

// LoadConfig will attempt to load the tweetstreem configuration file.
func (t *TweetStreem) LoadConfig() error {
	if err := tsViper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := tsViper.UnmarshalKey("config", t); err != nil {
		return fmt.Errorf("unmarshalling config failed: %w", err)
	}
	if err := t.parseTemplate(); err != nil {
		return err
	}
	return nil
}

// SaveConfig writes the current tweetstreem configuration to the configuration file.
func (t *TweetStreem) SaveConfig() error {
	if t.twitter != nil {
		cfg := t.twitter.Configuration()
		t.TwitterConfiguration = &cfg
	}

	tsViper.Set("config", t)
	savePath := filepath.Join(configPath, fmt.Sprint(configFile, ".", configFormat))
	if err := tsViper.WriteConfigAs(savePath); err != nil {
		return fmt.Errorf("saving config to %q failed: %w", savePath, err)
	}
	return nil
}
