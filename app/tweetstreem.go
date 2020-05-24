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
	"time"
)

const (
	TweetHistorySize = 100

	DefaultPort                  = 8080
	DefaultMentionHighlightColor = "blue"
	DefaultHashtagHighlightColor = "magenta"
)

type TweetStreem struct {
	TwitterConfiguration *TwitterConfiguration `json:"twitterConfiguration"`
	TweetTemplate        string                `json:"tweetTemplate"`
	TemplateOutputConfig OutputConfig          `json:"templateOutputConfig"`
	EnableApi            bool                  `json:"enableApi"`
	EnableClientLinks    bool                  `json:"enableClientLinks"`
	ApiPort              int                   `json:"apiPort"`
	ApiHost              string                `json:"apiHost"`
	AutoHome             bool                  `json:"autoHome"`

	api            *Api
	tweetTemplate  *template.Template
	twitter        *Twitter
	tweetHistory   map[int]*Tweet
	histLock       sync.Mutex
	lastTweetId    *int32
	inputCh        chan string
	printCh        chan string
	rpcCh          chan string
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
		ApiPort:              DefaultPort,
		TwitterConfiguration: &TwitterConfiguration{},
		TemplateOutputConfig: OutputConfig{
			MentionHighlightColor: DefaultMentionHighlightColor,
			HashtagHighlightColor: DefaultHashtagHighlightColor,
		},
		TweetTemplate: DefaultTweetTemplate,
		tweetHistory:  make(map[int]*Tweet),
		lastTweetId:   new(int32),
		inputCh:       make(chan string, 0),
		printCh:       make(chan string, 5),
		rpcCh:         make(chan string, 5),
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
	t.print(fmt.Sprintln(msg))
	t.print(request)
	confirmation := <-t.inputCh
	confIn, _ := SplitCommand(confirmation)
	if (defaultYes && confIn == "") || confIn == "y" || confIn == "yes" {
		return true
	}
	t.print(fmt.Sprintln(abort))
	return false
}

func (t *TweetStreem) RpcProcessCommand(args *Arguments, out *Output) error {
	var err error
	err = t.processCommand(true, args.Input)
	select {
	case o := <-t.rpcCh:
		out.Result = o
	case <-time.After(time.Second):
		// no output
	}
	return err
}

func (t *TweetStreem) processCommand(isRpc bool, input string) error {
	command, args := SplitCommand(input)
	var topErr error
	switch command {
	//case "c": // clear screen
	case "p", "pause": // pause the streem
		t.pause()
	case "r", "resume": // unpause the streem
		t.resume()
	case "v", "version": // show version
		t.print(version())
	case "o", "open":
		if n, ok := FirstNumber(args...); ok {
			idx, ok := FirstNumber(args[1:]...) // TODO:(smt) range check?
			if !ok {
				idx = 0
			}
			if tw, err := t.findTweet(n); err != nil {
				topErr = err
			} else {
				topErr = t.open(isRpc, tw, idx)
			}
		}
	case "b", "browse":
		if n, ok := FirstNumber(args...); ok {
			if tw, err := t.findTweet(n); err != nil {
				fmt.Sprintln(err)
			} else {
				err = t.browse(isRpc, tw)
			}
		}
	case "t", "tweet":
		message := strings.Join(args, " ")
		confirmMsg := fmt.Sprintln("tweet:", message)
		abortMsg := "tweet aborted\n"
		if t.confirmation(confirmMsg, abortMsg, true) {
			msg := t.tweet(message)
			t.print(msg)
		}
	case "reply":
		if n, ok := FirstNumber(args...); ok {
			msg := strings.Join(args[1:], " ")
			confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
			abortMsg := "reply aborted"
			if t.confirmation(confirmMsg, abortMsg, true) {
				m := t.reply(n, msg)
				t.print(m)
			}
		}
	case "cbreply":
		if n, ok := FirstNumber(args...); ok {
			if msg, err := ClipboardHelper.ReadAll(); err != nil {
				t.print(fmt.Sprintln("Error:", err))
			} else {
				confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
				abortMsg := "reply aborted"
				if t.confirmation(confirmMsg, abortMsg, true) {
					m := t.reply(n, msg)
					t.print(m)
				}
			}
		}
	case "urt", "unretweet":
		if n, ok := FirstNumber(args...); ok {
			m := t.unReTweet(n)
			t.print(m)
		}
	case "rt", "retweet":
		if n, ok := FirstNumber(args...); ok {
			msg := t.reTweet(n)
			t.print(msg)
		}
	case "ul", "unlike":
		if n, ok := FirstNumber(args...); ok {
			msg := t.unLike(n)
			t.print(msg)
		}
	case "li", "like":
		if n, ok := FirstNumber(args...); ok {
			msg := t.like(n)
			t.print(msg)
		}
	case "config":
		msg := t.config()
		t.print(msg)
	case "me":
		topErr = t.userTimeline(t.twitter.ScreenName())
	case "home":
		topErr = t.homeTimeline()
	case "h", "help":
		msg := t.help()
		t.print(msg)
	case "q", "quit", "exit":
		t.cancel()
	}
	return topErr
}

func (t *TweetStreem) consumeInput() {
	userPrompt := fmt.Sprintf("[@%s] ", t.twitter.ScreenName())
	for {
		t.print(Colors.Colorize("red", userPrompt))
		select {
		case <-t.ctx.Done():
			return
		case input := <-t.inputCh:
			if err := t.processCommand(false, input); err != nil {
				t.print(fmt.Sprintln("Error:", err))
			}
		}
	}
}

func (t *TweetStreem) rpcResponse(msg string) {
	if t.EnableApi {
		go func() { t.rpcCh <- msg }()
	}
}

func (t *TweetStreem) print(msg string) {
	go func() { t.printCh <- msg }()
}
func (t *TweetStreem) outputPrinter() {
	for m := range t.printCh {
		fmt.Print(m)
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

func (t *TweetStreem) resume() {
	t.print("resuming streem.\n")
	t.twitter.TogglePollerPaused(false)
}

func (t *TweetStreem) pause() {
	t.print("pausing streem.\n")
	t.twitter.TogglePollerPaused(true)
}

func (t *TweetStreem) timeLine(screenName string) error {
	conf := NewOaRequestConf()
	conf.Set("screen_name", screenName)
	tweets, err := t.twitter.UserTimeline(conf)
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) homeTimeline() error {
	tweets, err := t.twitter.HomeTimeline(NewOaRequestConf())
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) userTimeline(screenName string) error {
	cfg := NewOaRequestConf()
	cfg.Set("screen_name", screenName)
	tweets, err := t.twitter.UserTimeline(cfg)
	if err != nil {
		return err
	}
	t.clearHistory()
	t.echoTweets(tweets)
	return nil
}

func (t *TweetStreem) findTweet(id int) (*Tweet, error) {
	tw := t.getHistoryTweet(id)
	if tw == nil {
		return nil, fmt.Errorf("unknown tweet - id: %d", id)
	}
	return tw, nil
}

func (t *TweetStreem) browse(isRpc bool, tw *Tweet) error {
	if tw == nil {
		return fmt.Errorf("invalit tweet")
	}
	u := tw.HtmlLink()
	if t.EnableClientLinks && t.EnableApi && isRpc {
		t.rpcResponse(u)
	} else {
		t.print(fmt.Sprintln("opening in browser:", u))
		if err := OpenBrowser(u); err != nil {
			return fmt.Errorf("failed to open %q in browser: %w", u, err)
		}
	}
	return nil
}

// returns url that was opened or requestError
func (t *TweetStreem) open(isRpc bool, tw *Tweet, linkIdx int) error {
	if tw == nil {
		return fmt.Errorf("invalid tweet")
	}
	if linkIdx < 1 {
		linkIdx = 0
	}
	var u string
	ulist := tw.Links()
	if len(ulist) >= 0 && linkIdx < len(ulist) { // TODO: select url
		u = ulist[linkIdx]
	} else {
		return fmt.Errorf("could not find link for index: %d", linkIdx)
	}
	if t.EnableClientLinks && t.EnableApi && isRpc {
		t.rpcResponse(u)
	} else {
		t.print(fmt.Sprintln("opening in browser:", u))
		if err := OpenBrowser(u); err != nil {
			return fmt.Errorf("failed to open %q in browser: %w", u, err)
		}
	}
	return nil
}

func (t *TweetStreem) tweet(msg string) string {
	if len(msg) < 1 {
		return fmt.Sprintln("Some text is required to tweet")
	}
	if tw, err := t.twitter.UpdateStatus(msg, NewOaRequestConf()); err != nil {
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

	conf := NewOaRequestConf()
	conf.Set("in_reply_to_status_id", tw.IDStr)
	if tw, err := t.twitter.UpdateStatus(msg, conf); err != nil {
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
	if err := t.twitter.ReTweet(tw, NewOaRequestConf()); err != nil {
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
	if err := t.twitter.UnReTweet(tw, NewOaRequestConf()); err != nil {
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
	if err := t.twitter.Like(tw, NewOaRequestConf()); err != nil {
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
	if err := t.twitter.UnLike(tw, NewOaRequestConf()); err != nil {
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
			t.print(fmt.Sprintln("Error:", err))
		}
	}
}
