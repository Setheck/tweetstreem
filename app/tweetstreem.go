package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
)

const (
	TweetHistorySize = 100
)

type TweetStreem struct {
	*TwitterConfiguration `json:"twitterConfiguration"`
	TweetTemplate         string       `json:"tweetTemplate"`
	TemplateOutputConfig  OutputConfig `json:"templateOutputConfig"`
	EnableApi             bool         `json:"enableApi"`
	ApiPort               int          `json:"apiPort"`
	ApiHost               string       `json:"apiHost"`
	AutoHome              bool         `json:"autoHome"`

	api            *Api
	tweetTemplate  *template.Template
	twitter        *Twitter
	tweetHistory   map[int]*Tweet
	histLock       sync.Mutex
	lastTweetId    *int32
	inputCh        chan string
	nonInteractive bool
	ctx            context.Context
	cancel         context.CancelFunc
}

const DefaultTweetTemplate = `
{{ .UserName | color "cyan" }} {{ "@" | color "green" }}{{ .ScreenName | color "green" }} {{ .RelativeTweetTime | color "magenta" }}
id:{{ .Id }} {{ "rt:" | color "cyan" }}{{ .ReTweetCount | color "cyan" }} {{ "♥:" | color "red" }}{{ .FavoriteCount | color "red" }} via {{ .App | color "blue" }}
{{ .TweetText }}
`

func NewTweetStreem(ctx context.Context) *TweetStreem {
	if ctx == nil {
		ctx = context.Background()
	}
	twctx, cancel := context.WithCancel(ctx)
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
		inputCh:       make(chan string, 0),
		ctx:           twctx,
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

func (t *TweetStreem) initTwitter() error {
	templateHelpers := map[string]interface{}{
		"color": Colors.Colorize,
	}

	tpl, err := template.New("").
		Funcs(templateHelpers).
		Parse(t.TweetTemplate)

	if err != nil {
		return err
	}
	t.tweetTemplate = tpl
	t.twitter = NewTwitter(t.TwitterConfiguration)
	return t.twitter.Init()
}

func (t *TweetStreem) initApi() {
	t.nonInteractive = true
	t.api = NewApi(t.ctx, t.ApiPort, false) // pass in context so there is no need to call api.Stop()
	if err := t.api.Start(t); err != nil {
		panic(err) // TODO:
	}
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

func (t *TweetStreem) watchTerminalInput() {
	in := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			if in.Scan() {
				t.inputCh <- in.Text()
			}
		}
	}
}

func (t *TweetStreem) confirmation(msg, abort string, defaultYes bool) bool {
	if t.nonInteractive {
		return true
	}
	request := "please confirm "
	if defaultYes {
		request += "(Y/n):"
	} else {
		request += "(N/y):"
	}
	fmt.Println(msg)
	fmt.Print(request)
	confirmation := <-t.inputCh
	confIn, _ := SplitCommand(confirmation)
	if (defaultYes && confIn == "") || confIn == "y" || confIn == "yes" {
		return true
	}
	fmt.Println(abort)
	return false
}

func (t *TweetStreem) RpcProcessCommand(args *Arguments, out *string) error {
	var err error
	*out, err = t.processCommand(args.Input)
	return err
}

func (t *TweetStreem) processCommand(input string) (string, error) {
	command, args := SplitCommand(input)
	var output string
	var err error
	switch command {
	//case "c": // clear screen
	case "p", "pause": // pause the streem
		output = t.pause()
	case "r", "resume": // unpause the streem
		output = t.resume()
	case "v", "version": // show version
		output = version()
	case "o", "open":
		if n, ok := FirstNumber(args...); ok {
			idx, ok := FirstNumber(args[1:]...) // TODO:(smt) range check?
			if !ok {
				idx = 0
			}
			output = t.open(n, idx)
		}
	case "b", "browse":
		if n, ok := FirstNumber(args...); ok {
			output = t.browse(n)
		}
	case "t", "tweet":
		message := strings.Join(args, " ")
		confirmMsg := fmt.Sprintln("tweet:", message)
		abortMsg := "tweet aborted\n"
		if t.confirmation(confirmMsg, abortMsg, true) {
			output = t.tweet(message)
		}
	case "reply":
		if n, ok := FirstNumber(args...); ok {
			msg := strings.Join(args[1:], " ")
			confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
			abortMsg := "reply aborted"
			if t.confirmation(confirmMsg, abortMsg, true) {
				output = t.reply(n, msg)
			}
		}
	case "cbreply":
		if n, ok := FirstNumber(args...); ok {
			if msg, err := ClipboardHelper.ReadAll(); err != nil {
				fmt.Println("Error:", err)
			} else {
				confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
				abortMsg := "reply aborted"
				if t.confirmation(confirmMsg, abortMsg, true) {
					output = t.reply(n, msg)
				}
			}
		}
	case "urt", "unretweet":
		if n, ok := FirstNumber(args...); ok {
			output = t.unReTweet(n)
		}
	case "rt", "retweet":
		if n, ok := FirstNumber(args...); ok {
			output = t.reTweet(n)
		}
	case "ul", "unlike":
		if n, ok := FirstNumber(args...); ok {
			output = t.unLike(n)
		}
	case "li", "like":
		if n, ok := FirstNumber(args...); ok {
			output = t.like(n)
		}
	case "config":
		output = t.config()
	case "me":
		err = t.userTimeline(t.twitter.ScreenName())
	case "home":
		err = t.homeTimeline()
	case "h", "help":
		output = t.help()
	case "q", "quit", "exit":
		t.cancel()
	}
	return output, err
}

func (t *TweetStreem) consumeInput() {
	userPrompt := fmt.Sprintf("[@%s] ", t.twitter.ScreenName())
	for {
		fmt.Print(Colors.Colorize("red", userPrompt))
		select {
		case <-t.ctx.Done():
			return
		case input := <-t.inputCh:
			if output, err := t.processCommand(input); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Print(output)
			}
		}
	}
}

func (t *TweetStreem) help() string {
	return fmt.Sprintln("Options:\n" +
		"config - show the current configuration\n" +
		"p,pause - pause the stream\n" +
		"r,resume - resume the stream\n" +
		"v,version - print tweetstreem version\n" +
		"Select a tweet by id, eg: 'open 2'\n" +
		" o,open - open the link in the selected tweet (optionally provide 0 based index)\n" +
		" b,browse - open the selected tweet in a browser\n" +
		" rt,retweet - retweet the selected tweet\n" +
		" urt,unretweet - uretweet the selected tweet\n" +
		" li,like - like the selected tweet\n" +
		" ul,unlike - unlike the selected tweet\n" +
		" reply <id> <status> - reply to the tweet id (requires user mention, and confirmation)\n" +
		" cbreply <id> - reply to tweet id with clipboard contents (requires confirmation)\n" +
		"t,tweet <status> - create a new tweet and post (requires confirmation)\n" +
		"me - view your recent tweets\n" +
		"home - view your default timeline\n" +
		"h,help - this help menu\n" +
		"q,quit,exit - exit tweetstreem.")

}

func (t *TweetStreem) config() string {
	if b, err := json.Marshal(t); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintln(string(b))
	}
}

func (t *TweetStreem) resume() string {
	t.twitter.TogglePollerPaused(false)
	return fmt.Sprintln("resuming streem.")
}

func (t *TweetStreem) pause() string {
	t.twitter.TogglePollerPaused(true)
	return fmt.Sprintln("pausing streem.")
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

func (t *TweetStreem) browse(id int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	u := tw.HtmlLink()
	if err := OpenBrowser(u); err != nil {
		return fmt.Sprintln("Error:", err)
	}
	return fmt.Sprintln("opening tweet in browser:", u)
}

func (t *TweetStreem) open(id, linkIdx int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if linkIdx < 1 {
		linkIdx = 0
	}
	var u string
	ulist := tw.Links()
	if len(ulist) >= 0 && linkIdx < len(ulist) { // TODO: select url
		u = ulist[linkIdx]
	} else {
		return fmt.Sprintln("could not find link for index:", linkIdx)
	}
	if err := OpenBrowser(u); err != nil {
		return fmt.Sprintln("Error:", err)
	}
	return fmt.Sprintln("opening url in browser:", u)
}

func (t *TweetStreem) tweet(msg string) string {
	if len(msg) < 1 {
		return fmt.Sprintln("Some text is required to tweet")
	}
	if tw, err := t.twitter.UpdateStatus(msg, OaRequestConf{}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) reply(id int, msg string) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if !strings.Contains(msg, tw.User.ScreenName) {
		return fmt.Sprintf("reply must contain the original screen name [%s]\n", tw.User.ScreenName)
	}
	if tw, err := t.twitter.UpdateStatus(msg, OaRequestConf{
		InReplyToStatusId: tw.IDStr,
	}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) reTweet(id int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if err := t.twitter.ReTweet(tw, OaRequestConf{}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s retweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) unReTweet(id int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if err := t.twitter.UnReTweet(tw, OaRequestConf{}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s unretweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) like(id int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if err := t.twitter.Like(tw, OaRequestConf{}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s liked\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) unLike(id int) string {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return fmt.Sprintln("unknown tweet - id:", id)
	}
	if err := t.twitter.UnLike(tw, OaRequestConf{}); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s unliked\n", tw.User.ScreenName)
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
