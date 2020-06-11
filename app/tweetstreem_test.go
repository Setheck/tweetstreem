package app

import (
	"context"
	"fmt"
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
	err := tw.ParseTemplate()
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
		isRpc bool
		error bool
	}{
		{"help", "help", false, false},
		{"h", "h", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.processCommand(test.isRpc, test.input)
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
		isRpc bool
		error bool
	}{
		{"pause", "pause", false, false},
		{"p", "p", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			twitterMock := new(mocks.Client)
			twitterMock.On("SetPollerPaused", mock.AnythingOfType("bool")).Return()
			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			err := tw.processCommand(test.isRpc, test.input)
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
		isRpc bool
		error bool
	}{
		{"resume", "resume", false, false},
		{"r", "r", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			twitterMock := new(mocks.Client)
			twitterMock.On("SetPollerPaused", mock.AnythingOfType("bool")).Return()
			tw := NewTweetStreem(context.TODO())
			tw.twitter = twitterMock
			err := tw.processCommand(test.isRpc, test.input)
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
		isRpc bool
		error bool
	}{
		{"version", "version", false, false},
		{"v", "v", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tw := NewTweetStreem(context.TODO())
			err := tw.processCommand(test.isRpc, test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, version())
		})
	}
}

func TestTweetStreem_ProcessCommand_Open(t *testing.T) {
	obSave := openBrowser
	defer func() { openBrowser = obSave }()

	tests := []struct {
		name  string
		input string
		isRpc bool
		error bool
	}{
		{"open", "open 1", false, false},
		{"o", "o 1", false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &twitter.Tweet{
				Entities: twitter.Entities{
					Urls: []twitter.Url{{ExpandedUrl: "http://example.com"}},
				},
			}
			openBrowser = func(url string) error {
				t.Helper()
				assert.Equal(t, "http://example.com", url)
				return nil
			}
			tw := NewTweetStreem(context.TODO())
			tw.tweetHistory.Log(tweet)
			err := tw.processCommand(test.isRpc, test.input)
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
		isRpc bool
		error bool
	}{
		{"browse", "browse 1", false, false},
		{"b", "b 1", false, false},
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
			err := tw.processCommand(test.isRpc, test.input)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			verifyPrint(t, tw, fmt.Sprintf("opening in browser: %s\n", tweetUrl))
		})
	}
}

func TestTweetStreem_ProcessCommand_Reply(t *testing.T) {
	tests := []struct {
		name  string
		input string
		isRpc bool
		error bool
	}{
		{"reply", "reply 1 test hello", false, false},
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
			err := tw.processCommand(test.isRpc, test.input)
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
		isRpc bool
		error bool
	}{
		{"cbreply", "cbreply 1", false, false},
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
			err := tw.processCommand(test.isRpc, test.input)
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

func verifyPrint(t *testing.T, tw *TweetStreem, expected string) {
	t.Helper()
	select {
	case printed := <-tw.printCh:
		assert.Equal(t, expected, printed)
	case <-time.After(time.Millisecond * 10):
		t.Fail()
	}
}
