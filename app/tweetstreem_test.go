package app

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/Setheck/tweetstreem/util"
	mocks2 "github.com/Setheck/tweetstreem/util/mocks"

	"github.com/Setheck/tweetstreem/twitter"

	"github.com/stretchr/testify/mock"

	"github.com/Setheck/tweetstreem/twitter/mocks"

	"github.com/stretchr/testify/assert"
)

func TestNewTweetStreem(t *testing.T) {
	theCtx, cancel := context.WithCancel(context.TODO())
	tw := NewTweetStreem(theCtx)
	assert.Equal(t, DefaultPort, tw.ApiPort)
	assert.NotNil(t, tw.TwitterConfiguration)
	assert.Equal(t, DefaultMentionHighlightColor, tw.TemplateOutputConfig.MentionHighlightColor)
	assert.Equal(t, DefaultHashtagHighlightColor, tw.TemplateOutputConfig.HashtagHighlightColor)
	assert.True(t, tw.TemplateOutputConfig.Highlight)
	assert.Equal(t, DefaultTweetTemplate, tw.TweetTemplate)
	assert.NotNil(t, tw.tweetHistory)
	cancel()
	select {
	case <-tw.ctx.Done():
	case <-time.After(time.Millisecond * 10):
		t.Fail()
	}
}

func TestTweetStreem_ParseTemplate(t *testing.T) {
	tw := NewTweetStreem(context.TODO())
	assert.Nil(t, tw.tweetTemplate)
	err := tw.parseTemplate()
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, tw.tweetTemplate)
	assert.Equal(t, "tweetstreem", tw.tweetTemplate.Name())
}

func TestTweetStreem_ProcessCommand_Help(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"help", "help", false},
		{"h", "h", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, tw.help())
		})
	}
}

func TestTweetStreem_ProcessCommand_Pause(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"pause", "pause", false},
		{"p", "p", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			twitterMock := new(mocks.Client)
			twitterMock.On("SetPollerPaused", mock.AnythingOfType("bool")).Return()
			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPaused(t, tw, twitterMock)
		})
	}
}

func TestTweetStreem_ProcessCommand_Resume(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"resume", "resume", false},
		{"r", "r", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			twitterMock := new(mocks.Client)
			twitterMock.On("SetPollerPaused", mock.AnythingOfType("bool")).Return()
			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyResumed(t, tw, twitterMock)
		})
	}
}

func TestTweetStreem_ProcessCommand_Version(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"version", "version", false},
		{"v", "v", false},
		{"info", "info", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, appInfo())
		})
	}
}

func TestTweetStreem_ProcessCommand_Open(t *testing.T) {
	obSave := openBrowser
	defer func() { openBrowser = obSave }()

	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"open", "open 1", false},
		{"o", "o 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				Entities: twitter.Entities{
					Urls: []twitter.URL{{ExpandedURL: "http://example.com"}},
				},
			}
			openBrowser = func(url string) error {
				t.Helper()
				assert.Equal(t, "http://example.com", url)
				return nil
			}
			tw := NewTweetStreem(context.TODO())
			tw.tweetHistory.Log(tweet)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "opening in browser: http://example.com\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_Browse(t *testing.T) {
	obSave := openBrowser
	defer func() { openBrowser = obSave }()

	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"browse", "browse 1", false},
		{"b", "b 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}
			tweetUrl := fmt.Sprintf(twitter.TweetLinkUriTemplate, "test", "123")
			openBrowser = func(url string) error {
				t.Helper()
				assert.Equal(t, tweetUrl, url)
				return nil
			}
			tw := NewTweetStreem(context.TODO())
			tw.tweetHistory.Log(tweet)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, fmt.Sprintf("opening in browser: %s\n", tweetUrl))
		})
	}
}

func TestTweetStreem_ProcessCommand_Tweet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"tweet", "tweet test hello", false},
		{"t", "t test hello", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UpdateStatus",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("url.Values")).
				Return(&twitter.Tweet{IDStr: "0000"}, nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			sendConfirmation(t, tw, true)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "tweet: test hello\n\n")
			verifyPrint(t, tw, "please confirm (Y/n):")
			verifyPrint(t, tw, "tweet success! [0000]\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_Reply(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"reply", "reply 1 test hello", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UpdateStatus",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("url.Values")).
				Return(&twitter.Tweet{IDStr: "0000"}, nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			sendConfirmation(t, tw, true)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "reply to 1: test hello\n")
			verifyPrint(t, tw, "please confirm (Y/n):")
			verifyPrint(t, tw, "tweet success! [0000]\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_CBReply(t *testing.T) {
	cbSave := util.ClipboardHelper
	defer func() { util.ClipboardHelper = cbSave }()

	cbMock := new(mocks2.Clipper)
	cbMock.On("ReadAll").Return("test hello", nil)
	util.ClipboardHelper = cbMock

	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"cbreply", "cbreply 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UpdateStatus",
				mock.AnythingOfType("string"),
				mock.AnythingOfType("url.Values")).
				Return(&twitter.Tweet{IDStr: "0000"}, nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			sendConfirmation(t, tw, true)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "reply to 1: test hello\n")
			verifyPrint(t, tw, "please confirm (Y/n):")
			verifyPrint(t, tw, "tweet success! [0000]\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_UnRetweet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"unretweet", "unretweet 1", false},
		{"urt", "urt 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UnReTweet",
				tweet,
				mock.AnythingOfType("url.Values")).
				Return(nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "tweet by @test unretweeted\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_ReTweet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"retweet", "retweet 1", false},
		{"rt", "rt 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("ReTweet",
				tweet,
				mock.AnythingOfType("url.Values")).
				Return(nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			sendConfirmation(t, tw, true)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "tweet by @test retweeted\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_UnLike(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"unlike", "unlike 1", false},
		{"ul", "ul 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UnLike",
				tweet,
				mock.AnythingOfType("url.Values")).
				Return(nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "tweet by @test unliked\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_Like(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"like", "like 1", false},
		{"li", "li 1", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("Like",
				tweet,
				mock.AnythingOfType("url.Values")).
				Return(nil)

			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			tw.tweetHistory.Log(tweet)
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, "tweet by @test liked\n")
		})
	}
}

func TestTweetStreem_ProcessCommand_Config(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"config", "config", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, tw.config())
		})
	}
}

func TestTweetStreem_ProcessCommand_Me(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"me", "me", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
				Text:  "something",
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("UserTimeline",
				mock.AnythingOfType("url.Values")).
				Return([]*twitter.Tweet{tweet}, nil)

			twitterMock.On("ScreenName").
				Return("test")

			tw := NewTweetStreem(context.TODO())
			if err := tw.parseTemplate(); err != nil {
				assert.NoError(t, err)
			}
			tw.twitter = twitterMock

			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			expectedTwwet := "\n\x1b[36m\x1b[0m \x1b[32m@\x1b[0m\x1b[32mtest\x1b[0m \x1b[35m\x1b[0m\n" +
				"id:1 \x1b[36mrt:\x1b[0m\x1b[36m0\x1b[0m \x1b[31m♥:\x1b[0m\x1b[31m0\x1b[0m via \x1b[34m\x1b[0m\n" +
				"something\n"
			if runtime.GOOS == "windows" {
				expectedTwwet = "\n @test \nid:1 rt:0 ♥:0 via \nsomething\n"
			}
			verifyPrint(t, tw, expectedTwwet)
		})
	}
}

func TestTweetStreem_ProcessCommand_Home(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"home", "home", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				IDStr: "123",
				User:  twitter.User{ScreenName: "test"},
				Text:  "something",
			}

			twitterMock := new(mocks.Client)
			twitterMock.On("HomeTimeline",
				mock.AnythingOfType("url.Values")).
				Return([]*twitter.Tweet{tweet}, nil)

			twitterMock.On("ScreenName").
				Return("test")

			tw := NewTweetStreem(context.TODO())
			if err := tw.parseTemplate(); err != nil {
				assert.NoError(t, err)
			}
			tw.twitter = twitterMock

			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			expectedTweet := "\n\x1b[36m\x1b[0m \x1b[32m@\x1b[0m\x1b[32mtest\x1b[0m \x1b[35m\x1b[0m\n" +
				"id:1 \x1b[36mrt:\x1b[0m\x1b[36m0\x1b[0m \x1b[31m♥:\x1b[0m\x1b[31m0\x1b[0m via \x1b[34m\x1b[0m\n" +
				"something\n"
			if runtime.GOOS == "windows" {
				expectedTweet = "\n @test \nid:1 rt:0 ♥:0 via \nsomething\n"
			}
			verifyPrint(t, tw, expectedTweet)
		})
	}
}

func TestTweetStreem_ProcessCommand_Quit(t *testing.T) {
	tests := []struct {
		name  string
		input string
		error bool
	}{
		{"quit", "quit", false},
		{"q", "q", false},
		{"exit", "exit", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.ProcessCommand(test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			select {
			case <-tw.ctx.Done():
			case <-time.After(time.Millisecond * 10):
				t.Fail()
			}
		})
	}
}

func TestTweetStreem_InitApi(t *testing.T) {
	ts := NewTweetStreem(context.TODO())
	ts.testMode = true
	err := ts.InitApi()
	assert.NoError(t, err)
	assert.NotNil(t, ts.rpcListener)
	assert.True(t, ts.nonInteractive)
}

func TestTweetStreem_InitTwitter(t *testing.T) {
	ts := NewTweetStreem(context.TODO())
	ts.testMode = true
	err := ts.InitTwitter()
	assert.NoError(t, err)
	assert.NotNil(t, ts.twitter)
}

func TestTweetStreem_StartSubsystems(t *testing.T) {
	ts := NewTweetStreem(context.Background())
	ts.testMode = true

	buf := &bytes.Buffer{}
	stdin = buf
	buf.Write([]byte("v\n"))
	err := ts.StartSubsystems()
	assert.NoError(t, err)
	time.AfterFunc(time.Millisecond*20, ts.cancel)

	expectedPrompt := "\x1b[31m[@] \x1b[0m"
	if runtime.GOOS == "windows" {
		expectedPrompt = "[@] "
	}
	verifyPrint(t, ts, expectedPrompt)
	<-ts.ctx.Done()
	verifyPrint(t, ts, appInfo())
}

func sendConfirmation(t *testing.T, tw *TweetStreem, confirm bool) {
	t.Helper()
	confirmation := "n"
	if confirm {
		confirmation = "y"
	}
	go func() { tw.inputCh <- confirmation }()
}

func verifyPaused(t *testing.T, tw *TweetStreem, twitterMock *mocks.Client) {
	t.Helper()
	verifyPrint(t, tw, "pausing streem.\n")
	twitterMock.AssertCalled(t, "SetPollerPaused", true)
}

func verifyResumed(t *testing.T, tw *TweetStreem, twitterMock *mocks.Client) {
	t.Helper()
	verifyPrint(t, tw, "resuming streem.\n")
	twitterMock.AssertCalled(t, "SetPollerPaused", false)
}

var outputCh = make(chan string, 5)

func init() {
	// overwrite the output writer for test, this allows verifyPrint to work
	// in both situations, where the output printer is started, and when not.
	outputWriter = func(s string) {
		outputCh <- s
	}
}

func verifyPrint(t *testing.T, tw *TweetStreem, expected string) {
	t.Helper()
	select {
	case printed := <-outputCh:
		assert.Equal(t, expected, printed)
	case printed := <-tw.printCh:
		assert.Equal(t, expected, printed)
	case <-time.After(time.Millisecond * 10):
		t.Fail()
	}
}
