package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Setheck/tweetstreem/util"
	"github.com/spf13/viper"
)

const (
	ConfigFormat = "json"
)

var (
	ConfigPath = ""
	ConfigFile = ".tweetstreem"
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
	if ConfigPath == "" {
		ConfigPath = home
	}
	tsViper.SetConfigName(ConfigFile)
	tsViper.SetConfigType(ConfigFormat)
	tsViper.AddConfigPath(ConfigPath)
}

func (t *TweetStreem) LoadConfig() error {
	if err := tsViper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := tsViper.UnmarshalKey("config", t); err != nil {
		return fmt.Errorf("unmarshalling config failed: %w", err)
	}
	if err := t.ParseTemplate(); err != nil {
		return err
	}
	return nil
}

func (t *TweetStreem) SaveConfig() error {
	if t.twitter != nil {
		cfg := t.twitter.Configuration()
		t.TwitterConfiguration = &cfg
	}

	tsViper.Set("config", t)
	savePath := filepath.Join(ConfigPath, fmt.Sprint(ConfigFile, ".", ConfigFormat))
	if err := tsViper.WriteConfigAs(savePath); err != nil {
		return fmt.Errorf("saving config to %q failed: %w", savePath, err)
	}
	return nil
}
