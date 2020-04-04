package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

const Banner = `
~~~~~~~~~~~~~~~~
~~Tweet
~~   Streem
~~~~~~~~~~~~~~~~`

func init() {
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigFormat)
	viper.AddConfigPath("$HOME/")
	viper.AddConfigPath(".")
}

type TweetStreem struct {
	*TwitterConfiguration `json:"twitterConfiguration"`
	TweetTemplate         string `json:"tweetTemplate"`

	tweetTemplate *template.Template
	twitter       *Twitter
	ctx           context.Context
	cancel        context.CancelFunc
}

const DefaultTweetTemplate = `{{ .User }}
{{ .Status }}
{{ .Text }}

`

func NewTweetStreem() *TweetStreem {
	ctx, cancel := context.WithCancel(context.Background())
	return &TweetStreem{
		TwitterConfiguration: &TwitterConfiguration{},
		TweetTemplate:        DefaultTweetTemplate,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

// Run is the main entry point, returns result code
func (t *TweetStreem) Run() int {
	t.loadConfig()
	fmt.Printf("%s\npolling every: %s\n", Banner, t.PollTime.Truncate(time.Second).String())

	if err := t.InitTwitter(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	go t.echoOnPoll()
	go t.watchTerminal()
	go t.signalWatcher()
	<-t.ctx.Done()
	t.saveConfig()
	fmt.Println("\n'till next time o/ ")
	return 0
}

func (t *TweetStreem) InitTwitter() error {
	tpl, err := template.New("").Parse(t.TweetTemplate)
	if err != nil {
		return err
	}
	t.tweetTemplate = tpl
	t.twitter = NewTwitter(t.TwitterConfiguration)
	return t.twitter.Init()
}

func (t *TweetStreem) signalWatcher() {
	select {
	case <-t.ctx.Done():
	case <-Signal():
		t.cancel()
	}
}

func (t *TweetStreem) echoOnPoll() {
	for tweets := range t.twitter.startPoller() {
		t.EchoTweets(tweets)
	}
}

func (t *TweetStreem) handleInput() chan string {
	inputCh := make(chan string, 0)
	go func(ch chan string) {
		in := bufio.NewScanner(os.Stdin)
		for {
			select {
			case <-t.ctx.Done():
				close(inputCh)
				return
			default:
				if in.Scan() {
					input := in.Text()
					if len(input) > 0 {
						inputCh <- input
					}
				}
			}
		}
	}(inputCh)
	return inputCh
}

func (t *TweetStreem) watchTerminal() {
	inCh := t.handleInput()

	for {
		fmt.Print("> ")
		var err error
		select {
		case <-t.ctx.Done():
			return
		case input := <-inCh:
			switch strings.ToLower(input) {
			//case "c": // clear screen
			case "p": // pause the streem
				t.Pause()
			case "r": // unpause the streem
				t.Resume()
			case "v": // show version
				t.Version()
			case "h":
				fallthrough
			case "help":
				fmt.Println("Options:\n home - view your default timeline.\n exit - exit tweetstreem.\n help (h) - this help menu :D")
			case "q":
				fallthrough
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

func (t *TweetStreem) Version() {
	fmt.Println(Banner)
	fmt.Println("version:", Version)
	fmt.Println(" commit:", Commit)
	fmt.Println("  built:", Built)
}

func (t *TweetStreem) Resume() {
	fmt.Println("resuming streem.")
	t.twitter.TogglePollerPaused(false)
}

func (t *TweetStreem) Pause() {
	fmt.Println("pausing streem.")
	t.twitter.TogglePollerPaused(true)
}

func (t *TweetStreem) Home() error {
	tweets, err := t.twitter.HomeTimeline(GetConf{})
	if err != nil {
		return err
	}
	t.EchoTweets(tweets)
	return nil
}

func (t *TweetStreem) EchoTweets(tweets []*Tweet) {
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]
		_ = t.tweetTemplate.Execute(os.Stdout, struct {
			User   string
			Status string
			Text   string
		}{
			User:   tweet.UsrString(),
			Status: tweet.StatusString(),
			Text:   tweet.String(),
		})
		//fmt.Printf("%s\n%s\n%s\n\n", tweet.UsrString(), tweet.StatusString(), tweet.String())
	}
}

func (t *TweetStreem) loadConfig() {
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

func (t *TweetStreem) saveConfig() {
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
