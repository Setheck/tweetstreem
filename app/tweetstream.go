package app

import (
	"fmt"
	"log"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./.tweetstream/")

	AppToken = os.Getenv("APP_TOKEN")
	AppSecret = os.Getenv("APP_SECRET")
}

type TweetStream struct {
	*TwitterConfiguration `json:"twitterConfiguration"`
	twitter               *Twitter
}

func NewTweetStream() *TweetStream {
	return &TweetStream{}
}

func (t *TweetStream) Init() error {
	t.loadConfig()
	t.twitter = NewTwitter(t.TwitterConfiguration)
	err := t.twitter.Init()
	if err != nil {
		return err
	}
	go func() {
		for tweets := range t.twitter.StartPoller() {
			t.EchoTweets(tweets)
		}
	}()
	return nil
}

func (t *TweetStream) WatchTerminal() {
	for {
		input := prompt.Input(">>>", NilCompleter)
		var err error
		switch {
		case input == "exit":
			return
		case input == "home":
			err = t.Home()
		}
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func (t *TweetStream) Home() error {
	tweets, err := t.twitter.HomeTimeline(GetConf{})
	if err != nil {
		return err
	}
	t.EchoTweets(tweets)
	return nil
}

func (t *TweetStream) EchoTweets(tweets []*Tweet) {
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]
		fmt.Printf("%s\n%s\n%s\n\n", tweet.UsrString(), tweet.StatusString(), tweet.String())
	}
}

func (t *TweetStream) Stop() error {
	t.saveConfig()
	return nil
}

func (t *TweetStream) loadConfig() {
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

func (t *TweetStream) saveConfig() {
	viper.Set("config", t)
	if err := viper.WriteConfig(); err != nil {
		log.Println(err)
	}
}
