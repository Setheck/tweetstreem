package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	Version = "0.0.1"
	Commit  = "dev"
	Built   = "0"

	ConfigFile   = ".tweetstreem"
	ConfigFormat = "json"
)

func init() {
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigFormat)
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath(".")
}

type TweetStream struct {
	*TwitterConfiguration `json:"twitterConfiguration"`
	twitter               *Twitter
	ctx                   context.Context
	cancel                context.CancelFunc
}

func NewTweetStream() *TweetStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &TweetStream{
		TwitterConfiguration: &TwitterConfiguration{},
		ctx:                  ctx,
		cancel:               cancel,
	}
}

func (t *TweetStream) Run() error {
	t.loadConfig()
	fmt.Printf(`
~~~~~~~~~~~~~~~~
~~Tweet
~~   Streem
~~~~~~~~~~~~~~~~
polling every: %s
`, t.PollTime.Truncate(time.Second).String())
	t.twitter = NewTwitter(t.TwitterConfiguration)
	err := t.twitter.Init()
	if err != nil {
		return err
	}

	go t.echoOnPoll()
	go t.watchTerminal()
	go t.signalWatcher()
	<-t.ctx.Done()
	err = t.stop()
	fmt.Println("\n'till next time o/ ")
	return err
}

func (t *TweetStream) signalWatcher() {
	select {
	case <-t.ctx.Done():
	case <-Signal():
		t.cancel()
	}
}

func (t *TweetStream) echoOnPoll() {
	for tweets := range t.twitter.startPoller() {
		t.EchoTweets(tweets)
	}
}

func (t *TweetStream) handleInput() chan string {
	inputCh := make(chan string, 0)
	go func(ch chan string) {
		in := bufio.NewScanner(os.Stdin)
		for {
			if in.Scan() {
				input := in.Text()
				if len(input) > 0 {
					inputCh <- input
				}
			}
		}
	}(inputCh)
	return inputCh
}

func (t *TweetStream) watchTerminal() {
	inCh := t.handleInput()

	for {
		fmt.Print("> ")
		var err error
		select {
		case input := <-inCh:
			switch strings.ToLower(input) {
			case "h":
				fallthrough
			case "help":
				fmt.Println("Options:\n home - view your default timeline.\n exit - exit tweetstream.\n help (h) - this help menu :D")
			case "exit":
				t.cancel()
				return
			case "home":
				err = t.Home()
			}
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

func (t *TweetStream) stop() error {
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
