package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Setheck/tweetstreem/twitter"
	"github.com/Setheck/tweetstreem/util"
)

const (
	DefaultPort                  = 8080
	DefaultMentionHighlightColor = "blue"
	DefaultHashtagHighlightColor = "magenta"
)

type TweetStreem struct {
	TwitterConfiguration *twitter.Configuration `json:"twitterConfiguration"`
	TweetTemplate        string                 `json:"tweetTemplate"`
	TemplateOutputConfig twitter.OutputConfig   `json:"templateOutputConfig"`
	EnableApi            bool                   `json:"enableApi"`
	EnableClientLinks    bool                   `json:"enableClientLinks"`
	ApiPort              int                    `json:"apiPort"`
	ApiHost              string                 `json:"apiHost"`
	AutoHome             bool                   `json:"autoHome"`

	api            Api
	tweetTemplate  *template.Template
	twitter        twitter.Client
	tweetHistory   *History
	inputCh        chan string
	printCh        chan string
	rpcCh          chan string
	nonInteractive bool
	ctx            context.Context
	cancel         context.CancelFunc
	testMode       bool // disable twitter auth, and server start for testing
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
		TwitterConfiguration: &twitter.Configuration{},
		TemplateOutputConfig: twitter.OutputConfig{
			MentionHighlightColor: DefaultMentionHighlightColor,
			HashtagHighlightColor: DefaultHashtagHighlightColor,
			Highlight:             true,
		},
		TweetTemplate: DefaultTweetTemplate,
		tweetHistory:  NewHistory(),
		inputCh:       make(chan string),
		printCh:       make(chan string, 5),
		rpcCh:         make(chan string, 5),
		ctx:           twctx,
		cancel:        cancel,
	}
}

func (t *TweetStreem) getHistoryTweet(id int) (*twitter.Tweet, error) {
	if tw, ok := t.tweetHistory.Last(id); ok {
		return tw.(*twitter.Tweet), nil
	}
	return nil, fmt.Errorf("unknown tweet - id:%d", id)
}

func (t *TweetStreem) RemoteCall() error {
	client := NewRemoteClient(t, fmt.Sprintf("%s:%d", t.ApiHost, t.ApiPort))
	input := strings.Join(flag.Args(), " ")
	if err := client.RpcCall(input); err != nil {
		return err
	}
	return nil
}

func (t *TweetStreem) parseTemplate() error {
	templateHelpers := map[string]interface{}{
		"color": util.Colors.Colorize,
	}
	tpl, err := template.New("tweetstreem").
		Funcs(templateHelpers).
		Parse(t.TweetTemplate)
	if err != nil {
		return err
	}

	t.tweetTemplate = tpl
	return nil
}

func (t *TweetStreem) SetNonInteractive(b bool) {
	t.nonInteractive = b
}

func (t *TweetStreem) InitApi() error {
	t.SetNonInteractive(true)
	t.api = NewApi(t.ctx, t.ApiPort, false)
	if !t.testMode {
		return t.api.Start(t) // pass in context so there is no need to call api.Stop()
	}
	return nil
}

func (t *TweetStreem) InitTwitter() error {
	t.twitter = twitter.NewDefaultClient(*t.TwitterConfiguration)
	if !t.testMode {
		if err := t.twitter.Authorize(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TweetStreem) WaitForDone() {
	select {
	case <-t.ctx.Done():
	case <-util.Signal():
		t.cancel()
	}
}

func (t *TweetStreem) pollAndEcho() {
	tweetCh := make(chan []*twitter.Tweet)
	t.twitter.StartPoller(tweetCh)
	for tweets := range tweetCh {
		t.PrintTweets(tweets)
	}
}

var Stdin io.Reader = os.Stdin

func (t *TweetStreem) watchStdin() {
	in := bufio.NewScanner(Stdin)
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

func (t *TweetStreem) userConfirmation(msg, abort string, defaultYes bool) bool {
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
	confIn, _ := util.SplitCommand(confirmation)
	if (defaultYes && confIn == "") || confIn == "y" || confIn == "yes" {
		return true
	}
	t.print(fmt.Sprintln(abort))
	return false
}

func (t *TweetStreem) RpcProcessCommand(args *Arguments, out *Output) error {
	var err = t.processCommand(true, args.Input)
	select {
	case o := <-t.rpcCh:
		out.Result = o
	case <-time.After(time.Second):
		// no output
	}
	return err
}

func (t *TweetStreem) ProcessCommand(input string) error {
	return t.processCommand(false, input)
}

func (t *TweetStreem) processCommand(isRpc bool, input string) error {
	command, args := util.SplitCommand(input)
	switch command {
	//case "c": // clear screen
	case "p", "pause": // pause the streem
		t.pause()
	case "r", "resume": // unpause the streem
		t.resume()
	case "v", "version", "info": // show appInfo
		t.print(appInfo())
	case "o", "open":
		return t.commandOpen(isRpc, args...)
	case "b", "browse":
		return t.commandBrowse(isRpc, args...)
	case "t", "tweet":
		t.commandTweet(isRpc, args...)
	case "reply":
		t.commandReply(args...)
	case "cbreply":
		t.clipBoardReply(args...)
	case "urt", "unretweet":
		t.commandUnReTweet(args...)
	case "rt", "retweet":
		t.commandRetweet(args...)
	case "ul", "unlike":
		t.commandUnLike(args...)
	case "li", "like":
		t.commandLike(args...)
	case "config":
		t.print(t.config())
	case "me":
		return t.userTimeline(t.twitter.ScreenName())
	case "home":
		return t.homeTimeline()
	case "h", "help":
		t.print(t.help())
	case "q", "quit", "exit":
		t.cancel()
	}
	return nil
}

func (t *TweetStreem) StartSubsystems() error {
	if err := t.InitTwitter(); err != nil {
		return err
	}

	go t.consumeInput()
	go t.outputPrinter()
	go t.pollAndEcho()
	go t.watchStdin()

	if t.EnableApi {
		fmt.Println("api server enabled on port:", t.ApiPort)
		if err := t.InitApi(); err != nil {
			return err
		}
	}
	if t.AutoHome {
		_ = t.homeTimeline()
	}
	return nil
}

func (t *TweetStreem) consumeInput() {
	userPrompt := fmt.Sprintf("[@%s] ", t.twitter.ScreenName())
	for {
		t.print(util.Colors.Colorize("red", userPrompt))
		select {
		case <-t.ctx.Done():
			return
		case input := <-t.inputCh:
			if err := t.ProcessCommand(input); err != nil {
				t.print(fmt.Sprintln("Error:", err))
			}
		}
	}
}

func (t *TweetStreem) rpcResponse(msg string) {
	if t.EnableApi {
		select {
		case t.rpcCh <- msg:
		case <-time.After(time.Millisecond * 500):
			fmt.Println("error dropped rpcResponse:", msg)
		}
	}
}

func (t *TweetStreem) print(msg string) {
	select {
	case t.printCh <- msg:
	case <-time.After(time.Millisecond * 500):
		fmt.Println("error dropped print:", msg)
	}
}

var outputWriter = func(s string) { fmt.Print(s) }

func (t *TweetStreem) outputPrinter() {
	for m := range t.printCh {
		outputWriter(m)
	}
}

func (t *TweetStreem) help() string {
	return fmt.Sprintln("Options:\n" +
		"config - show the current configuration\n" +
		"p,pause - pause the stream\n" +
		"r,resume - resume the stream\n" +
		"v,appInfo - print tweetstreem appInfo\n" +
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
	t.twitter.SetPollerPaused(false)
}

func (t *TweetStreem) pause() {
	t.print("pausing streem.\n")
	t.twitter.SetPollerPaused(true)
}

func (t *TweetStreem) timeLine(screenName string) error {
	conf := twitter.NewURLValues()
	conf.Set("screen_name", screenName)
	tweets, err := t.twitter.UserTimeline(conf)
	if err != nil {
		return err
	}
	t.tweetHistory.Clear()
	t.PrintTweets(tweets)
	return nil
}

func (t *TweetStreem) homeTimeline() error {
	tweets, err := t.twitter.HomeTimeline(twitter.NewURLValues())
	if err != nil {
		return err
	}
	t.tweetHistory.Clear()
	t.PrintTweets(tweets)
	return nil
}

func (t *TweetStreem) userTimeline(screenName string) error {
	cfg := twitter.NewURLValues()
	cfg.Set("screen_name", screenName)
	tweets, err := t.twitter.UserTimeline(cfg)
	if err != nil {
		return err
	}
	t.tweetHistory.Clear()
	t.PrintTweets(tweets)
	return nil
}

func (t *TweetStreem) findTweet(id int) (*twitter.Tweet, error) {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return nil, err
	}
	return tw, nil
}

func (t *TweetStreem) commandBrowse(isRpc bool, args ...string) error {
	if n, ok := util.FirstNumber(args...); ok {
		if tw, err := t.findTweet(n); err != nil {
			t.print(fmt.Sprintln(err))
		} else {
			return t.browse(isRpc, tw)
		}
	}
	return fmt.Errorf("invalid tweet id")
}

// test point
var openBrowser = util.OpenBrowser

func (t *TweetStreem) browse(isRpc bool, tw *twitter.Tweet) error {
	if tw == nil {
		return fmt.Errorf("invalit tweet")
	}
	u := tw.HTMLLink()
	if t.EnableClientLinks && t.EnableApi && isRpc {
		t.rpcResponse(u)
	} else {
		t.print(fmt.Sprintln("opening in browser:", u))
		if err := openBrowser(u); err != nil {
			return fmt.Errorf("failed to open %q in browser: %w", u, err)
		}
	}
	return nil
}

func (t *TweetStreem) commandOpen(isRpc bool, args ...string) error {
	if n, ok := util.FirstNumber(args...); ok {
		idx, ok := util.FirstNumber(args[1:]...) // TODO:(smt) range check?
		if !ok {
			idx = 0
		}
		if tw, err := t.findTweet(n); err != nil {
			return err
		} else {
			return t.open(isRpc, tw, idx)
		}
	}
	return fmt.Errorf("invalid tweet id")
}

// returns url that was opened or requestError
func (t *TweetStreem) open(isRpc bool, tw *twitter.Tweet, linkIdx int) error {
	if tw == nil {
		return fmt.Errorf("invalid tweet")
	}
	if linkIdx < 1 {
		linkIdx = 0
	}
	ulist := tw.Links()
	if linkIdx >= len(ulist) {
		return fmt.Errorf("could not find link for index: %d", linkIdx)
	}
	u := ulist[linkIdx]
	if t.EnableClientLinks && t.EnableApi && isRpc {
		t.rpcResponse(u)
	} else {
		t.print(fmt.Sprintln("opening in browser:", u))
		if err := openBrowser(u); err != nil {
			return fmt.Errorf("failed to open %q in browser: %w", u, err)
		}
	}
	return nil
}

func (t *TweetStreem) commandTweet(isRpc bool, args ...string) {
	message := strings.Join(args, " ")
	confirmMsg := fmt.Sprintln("tweet:", message)
	abortMsg := "tweet aborted\n"
	if t.userConfirmation(confirmMsg, abortMsg, true) {
		msg := t.tweet(message)
		t.print(msg)
	}
}

func (t *TweetStreem) tweet(msg string) string {
	if len(msg) < 1 {
		return fmt.Sprintln("some text is required to tweet")
	}
	if tw, err := t.twitter.UpdateStatus(msg, twitter.NewURLValues()); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) commandReply(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		msg := strings.Join(args[1:], " ")
		confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
		abortMsg := "reply aborted"
		if t.userConfirmation(confirmMsg, abortMsg, true) {
			m := t.reply(n, msg)
			t.print(m)
		}
	}
}

func (t *TweetStreem) clipBoardReply(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		if msg, err := util.ClipboardHelper.ReadAll(); err != nil {
			t.print(fmt.Sprintln("Error:", err))
		} else {
			confirmMsg := fmt.Sprintf("reply to %d: %s", n, msg)
			abortMsg := "reply aborted"
			if t.userConfirmation(confirmMsg, abortMsg, true) {
				m := t.reply(n, msg)
				t.print(m)
			}
		}
	}
}

func (t *TweetStreem) reply(id int, msg string) string {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return err.Error()
	}
	if !strings.Contains(msg, tw.User.ScreenName) {
		return fmt.Sprintf("reply must contain the original screen name [%s]\n", tw.User.ScreenName)
	}

	conf := twitter.NewURLValues()
	conf.Set("in_reply_to_status_id", tw.IDStr)
	if tw, err := t.twitter.UpdateStatus(msg, conf); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet success! [%s]\n", tw.IDStr)
	}
}

func (t *TweetStreem) commandRetweet(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		msg := t.reTweet(n)
		t.print(msg)
	}
}

func (t *TweetStreem) reTweet(id int) string {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return err.Error()
	}
	if err := t.twitter.ReTweet(tw, twitter.NewURLValues()); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s retweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) commandUnReTweet(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		m := t.unReTweet(n)
		t.print(m)
	}
}

func (t *TweetStreem) unReTweet(id int) string {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return err.Error()
	}
	if err := t.twitter.UnReTweet(tw, twitter.NewURLValues()); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s unretweeted\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) commandLike(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		msg := t.like(n)
		t.print(msg)
	}
}

func (t *TweetStreem) like(id int) string {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return err.Error()
	}
	if err := t.twitter.Like(tw, twitter.NewURLValues()); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s liked\n", tw.User.ScreenName)
	}
}

func (t *TweetStreem) commandUnLike(args ...string) {
	if n, ok := util.FirstNumber(args...); ok {
		msg := t.unLike(n)
		t.print(msg)
	}
}

func (t *TweetStreem) unLike(id int) string {
	tw, err := t.getHistoryTweet(id)
	if err != nil {
		return err.Error()
	}
	if err := t.twitter.UnLike(tw, twitter.NewURLValues()); err != nil {
		return fmt.Sprintln("Error:", err)
	} else {
		return fmt.Sprintf("tweet by @%s unliked\n", tw.User.ScreenName)
	}
}

var Stdout = os.Stdout

func (t *TweetStreem) PrintTweets(tweets []*twitter.Tweet) {
	for i := len(tweets) - 1; i >= 0; i-- {
		tweet := tweets[i]
		t.tweetHistory.Log(tweet)
		buf := new(bytes.Buffer)
		if err := t.tweetTemplate.Execute(buf, struct {
			Id int
			twitter.TweetTemplateOutput
		}{
			Id:                  t.tweetHistory.LastIdx(),
			TweetTemplateOutput: tweet.TemplateOutput(t.TemplateOutputConfig),
		}); err != nil {
			t.print(fmt.Sprintln("Error:", err))
		} else {
			t.print(buf.String())
		}
	}
}
