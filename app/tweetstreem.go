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
	viper.AddConfigPath("$HOME/") // TODO:(smt) how does this work on windows.
	viper.AddConfigPath(".")
}

type TweetStreem struct {
	*TwitterConfiguration `json:"twitterConfiguration"`
	TweetTemplate         string       `json:"tweetTemplate"`
	TemplateOutputConfig  OutputConfig `json:"templateOutputConfig"`
	EnableApi             bool         `json:"enableApi"`
	ApiPort               int          `json:"apiPort"`
	AutoHome              bool         `json:"autoHome"`

	api           *Api
	tweetTemplate *template.Template
	twitter       *Twitter
	tweetHistory  map[int]*Tweet
	histLock      sync.Mutex
	lastTweetId   *int32
	ctx           context.Context
	cancel        context.CancelFunc
}

const DefaultTweetTemplate = `
{{ .UserName | color "cyan" }} {{ "@" | color "green" }}{{ .ScreenName | color "green" }} {{ .RelativeTweetTime | color "magenta" }}
id:{{ .Id }} {{ "rt:" | color "cyan" }}{{ .ReTweetCount | color "cyan" }} {{ "♥:" | color "red" }}{{ .FavoriteCount | color "red" }} via {{ .App | color "blue" }}
{{ .TweetText }}
`

func NewTweetStreem() *TweetStreem {
	ctx, cancel := context.WithCancel(context.Background())
	return &TweetStreem{
		ApiPort:              8080,
		TwitterConfiguration: &TwitterConfiguration{},
		TemplateOutputConfig: OutputConfig{
			MentionHighlightColor: "blue",
			HashtagHighlightColor: "magenta",
		},
		TweetTemplate: DefaultTweetTemplate,
		tweetHistory:  make(map[int]*Tweet),
		lastTweetId:   new(int32),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (t *TweetStreem) logTweet(id int, tweet *Tweet) {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	if id > TweetHistorySize {
		delete(t.tweetHistory, id-TweetHistorySize)
	}
	t.tweetHistory[id] = tweet
}

func (t *TweetStreem) getHistoryTweet(id int) *Tweet {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	return t.tweetHistory[id]
}

func (t *TweetStreem) clearHistory() {
	t.histLock.Lock()
	defer t.histLock.Unlock()
	t.tweetHistory = make(map[int]*Tweet)
	atomic.StoreInt32(t.lastTweetId, 0)
}

// Run is the main entry point, returns result code
func (t *TweetStreem) Run() int {
	t.loadConfig()
	fmt.Printf("%s\npolling every: %s\n", Banner, t.PollTime.Truncate(time.Second).String())

	if err := t.initTwitter(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	if t.EnableApi {
		t.initApi()
	}
	go t.echoOnPoll()
	go t.watchTerminal()
	go t.signalWatcher()

	if t.AutoHome {
		t.homeTimeline()
	}
	<-t.ctx.Done()
	t.saveConfig()
	fmt.Println("\n'till next time o/ ")
	return 0
}

func (t *TweetStreem) initTwitter() error {
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

func (t *TweetStreem) initApi() {
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
		t.echoTweets(tweets)
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
	userPrompt := fmt.Sprintf("[@%s] ", t.twitter.ScreenName())
	for {
		fmt.Print(Colors.Colorize("red", userPrompt))
		var err error
		select {
		case <-t.ctx.Done():
			return
		case input := <-inCh:
			command, args := SplitCommand(input)
			switch command {
			//case "c": // clear screen
			case "p", "pause": // pause the streem
				t.pause()
			case "r", "resume": // unpause the streem
				t.resume()
			case "v", "version": // show version
				t.Version()
			case "o", "open":
				if n, ok := FirstNumber(args...); ok {
					t.open(n)
				}
			case "b", "browse":
				if n, ok := FirstNumber(args...); ok {
					t.browse(n)
				}
			//case "t", "tweet":
			// TODO
			//case "reply":
			// TODO
			//case "cbreply": // TODO: confirmation
			//	if n, ok := FirstNumber(args...); ok {
			//		if msg, err := ClipboardHelper.ReadAll(); err != nil {
			//			fmt.Println("Error:", err)
			//		} else {
			//			t.reply(n, msg)
			//		}
			//	}
			case "urt", "unretweet":
				if n, ok := FirstNumber(args...); ok {
					t.unReTweet(n)
				}
			case "rt", "retweet":
				if n, ok := FirstNumber(args...); ok {
					t.reTweet(n)
				}
			case "ul", "unlike":
				if n, ok := FirstNumber(args...); ok {
					t.unLike(n)
				}
			case "li", "like":
				if n, ok := FirstNumber(args...); ok {
					t.like(n)
				}
			case "config":
				t.config()
			case "me":
				err = t.userTimeline(t.twitter.ScreenName())
			case "home":
				err = t.homeTimeline()
			case "h", "help":
				t.help()
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

func (t *TweetStreem) help() {
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
		"me - view your recent tweets\n" +
		"home - view your default timeline\n" +
		"h help - this help menu\n" +
		"q,quit,exit - exit tweetstreem.\n" +
		"help (h) - this help menu :D")

}

func (t *TweetStreem) config() {
	if b, err := json.Marshal(t); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(string(b))
	}

}

func (t *TweetStreem) Version() {
	fmt.Println(Banner)
	fmt.Println("version:", Version)
	fmt.Println(" commit:", Commit)
	fmt.Println("  built:", Built)
}

func (t *TweetStreem) resume() {
	fmt.Println("resuming streem.")
	t.twitter.TogglePollerPaused(false)
}

func (t *TweetStreem) pause() {
	fmt.Println("pausing streem.")
	t.twitter.TogglePollerPaused(true)
}

func (t *TweetStreem) timeLine(screenName string) error {
	conf := OaRequestConf{screenName: screenName}
	tweets, err := t.twitter.UserTimeline(conf)
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) homeTimeline() error {
	tweets, err := t.twitter.HomeTimeline(OaRequestConf{})
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) userTimeline(screenName string) error {
	tweets, err := t.twitter.UserTimeline(OaRequestConf{screenName: screenName})
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) browse(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if err := OpenBrowser(tw.HtmlLink()); err != nil {
		fmt.Println("Error:", err)
	}
}

func (t *TweetStreem) open(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	ulist := tw.ExpandedUrls()
	if len(ulist) > 0 { // TODO: select url
		if err := OpenBrowser(ulist[0]); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("tweet contains no links")
	}
}

func (t *TweetStreem) tweet(msg string) {
	if len(msg) < 1 {
		fmt.Println("Some text is required to tweet")
		return
	}
	if tw, err := t.twitter.UpdateStatus(msg, OaRequestConf{}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) reply(id int, msg string) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if !strings.Contains(msg, tw.User.ScreenName) {
		fmt.Printf("reply must contain the original screen name [%s]\n", tw.User.ScreenName)
		return
	}
	if tw, err := t.twitter.UpdateStatus(msg, OaRequestConf{
		InReplyToStatusId: tw.IDStr,
	}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) reTweet(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if err := t.twitter.ReTweet(tw, OaRequestConf{}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet by @%s retweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) unReTweet(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if err := t.twitter.UnReTweet(tw, OaRequestConf{}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet by @%s unretweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) like(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if err := t.twitter.Like(tw, OaRequestConf{}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet by @%s liked\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) unLike(id int) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		fmt.Println("unknown tweet - id:", id)
		return
	}
	if err := t.twitter.UnLike(tw, OaRequestConf{}); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("tweet by @%s unliked\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) echoTweets(tweets []*Tweet) {
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]
		atomic.AddInt32(t.lastTweetId, 1)
		t.logTweet(int(*t.lastTweetId), tweet)
		if err := t.tweetTemplate.Execute(os.Stdout, struct {
			Id int
			TweetTemplateOutput
		}{
			Id:                  int(*t.lastTweetId),
			TweetTemplateOutput: tweet.TemplateOutput(t.TemplateOutputConfig),
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
