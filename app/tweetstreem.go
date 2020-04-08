package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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

const (
	TweetHistorySize = 100
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
	EnableApi             bool   `json:"enableApi"`
	ApiPort               int    `json:"apiPort"`
	AutoHome              bool   `json:"autoHome"`

	api           *Api
	tweetTemplate *template.Template
	twitter       *Twitter
	tweetHistory  map[int]*Tweet
	histLock      sync.Mutex
	lastTweetId   *int32
	ctx           context.Context
	cancel        context.CancelFunc
}

const DefaultTweetTemplate = `{{ .UserName | color "cyan" }} {{ .ScreenName | color "green" }} {{ .RelativeTweetTime | color "magenta" }}
id:{{ .Id }} {{ "rt:" | color "cyan" }}{{ .ReTweetCount | color "cyan" }} {{ "♥:" | color "red" }}{{ .FavoriteCount | color "red" }} via {{ .App | color "blue" }}
{{ .TweetText }}

`

func NewTweetStreem() *TweetStreem {
	ctx, cancel := context.WithCancel(context.Background())
	return &TweetStreem{
		ApiPort:              8080,
		TwitterConfiguration: &TwitterConfiguration{},
		TweetTemplate:        DefaultTweetTemplate,
		tweetHistory:         make(map[int]*Tweet),
		lastTweetId:          new(int32),
		ctx:                  ctx,
		cancel:               cancel,
	}
}

func (t *TweetStreem) LogTweet(id int, tweet *Tweet) {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	if id > TweetHistorySize {
		delete(t.tweetHistory, id-TweetHistorySize)
	}
	t.tweetHistory[id] = tweet
}

func (t *TweetStreem) GetHistoryTweet(id int) *Tweet {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	return t.tweetHistory[id]
}

func (t *TweetStreem) ClearHistory() {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	t.tweetHistory = make(map[int]*Tweet)
	atomic.StoreInt32(t.lastTweetId, 0)
}

// Run is the main entry point, returns result code
func (t *TweetStreem) Run() int {
	t.loadConfig()
	fmt.Printf("%s\npolling every: %s\n", Banner, t.PollTime.Truncate(time.Second).String())

	if err := t.InitTwitter(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	if t.EnableApi {
		t.InitApi()
	}
	go t.echoOnPoll()
	go t.watchTerminal()
	go t.signalWatcher()

	if t.AutoHome {
		t.Home()
	}
	<-t.ctx.Done()
	t.saveConfig()
	fmt.Println("\n'till next time o/ ")
	return 0
}

func (t *TweetStreem) InitTwitter() error {
	tpl, err := template.New("").Funcs(
		map[string]interface{}{
			"color": Colors.Colorize,
		}).Parse(t.TweetTemplate)
	if err != nil {
		return err
	}
	t.tweetTemplate = tpl
	t.twitter = NewTwitter(t.TwitterConfiguration)
	return t.twitter.Init()
}

func (t *TweetStreem) InitApi() {
	t.api = NewApi(t.ctx, t.ApiPort) // pass in context so there is no need to call api.Stop()
	t.api.Start()
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
			command, args := t.SplitCommand(input)
			switch command {
			//case "c": // clear screen
			case "p", "pause": // pause the streem
				t.Pause()
			case "r", "resume": // unpause the streem
				t.Resume()
			case "v", "version": // show version
				t.Version()
			case "o", "open":
				if n, ok := FirstNumber(args...); ok {
					t.Open(n)
				}
			case "b", "browse":
				if n, ok := FirstNumber(args...); ok {
					t.Browse(n)
				}
			case "urt", "unretweet":
				if n, ok := FirstNumber(args...); ok {
					t.UnReTweet(n)
				}
			case "rt", "retweet":
				if n, ok := FirstNumber(args...); ok {
					t.ReTweet(n)
				}
			case "ul", "unlike":
				if n, ok := FirstNumber(args...); ok {
					t.UnLike(n)
				}
			case "li", "like":
				if n, ok := FirstNumber(args...); ok {
					t.Like(n)
				}
			case "config":
				t.Config()
			case "home":
				err = t.Home()
			case "h", "help":
				t.Help()
			case "q", "quit", "exit":
				t.cancel()
				return
			}
		}
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func (t *TweetStreem) Help() {
	fmt.Println("Options:\n" +
		"config - show the current configuration\n" +
		"p,pause - pause the stream\n" +
		"r,resume - resume the stream\n" +
		"v,version - print tweetstreem version\n" +
		"Select a tweet by id, eg: 'open 2'\n" +
		" o,open - open the link in the selected tweet\n" +
		" b,browse - open the selected tweet in a browser\n" +
		" rt,retweet - retweet the selected tweet\n" +
		" urt,unretweet - uretweet the selected tweet\n" +
		" li,like - like the selected tweet\n" +
		" ul,unlike - unlike the selected tweet\n" +
		"home - view your default timeline\n" +
		"h help - this help menu\n" +
		"q,quit,exit - exit tweetstreem.\n" +
		"help (h) - this help menu :D")

}

func (t *TweetStreem) Config() {
	if b, err := json.Marshal(t); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(string(b))
	}

}

// SplitCommand takes a string and resturns command and arguments
func (t *TweetStreem) SplitCommand(str string) (string, []string) {
	str = strings.ToLower(str)
	str = strings.TrimSpace(str)
	split := strings.Split(str, " ")
	if len(split) > 1 {
		return split[0], split[1:]
	}
	if len(split) > 0 {
		return split[0], nil
	}
	return "", nil
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
	tweets, err := t.twitter.HomeTimeline(OaRequestConf{})
	if err != nil {
		return err
	}
	t.ClearHistory()
	t.EchoTweets(tweets)
	return nil
}

func (t *TweetStreem) Browse(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		if err := OpenBrowser(tw.HtmlLink()); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) Open(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		fmt.Println("TODO: Open Tweet", tw)
		//OpenBrowser(tw.)
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) ReTweet(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		if err := t.twitter.ReTweet(tw, OaRequestConf{}); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) UnReTweet(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		if err := t.twitter.UnReTweet(tw, OaRequestConf{}); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) Like(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		if err := t.twitter.Like(tw, OaRequestConf{}); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) UnLike(id int) {
	if tw := t.GetHistoryTweet(id); tw != nil {
		if err := t.twitter.UnLike(tw, OaRequestConf{}); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("unknown tweet - id:", id)
	}
}

func (t *TweetStreem) EchoTweets(tweets []*Tweet) {
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]
		atomic.AddInt32(t.lastTweetId, 1)
		t.LogTweet(int(*t.lastTweetId), tweet)
		if err := t.tweetTemplate.Execute(os.Stdout, struct {
			Id int
			TweetTemplateOutput
		}{
			Id:                  int(*t.lastTweetId),
			TweetTemplateOutput: tweet.TemplateOutput(),
		}); err != nil {
			fmt.Println("Error:", err)
		}
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
