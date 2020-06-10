package app

import (
	"context"
	"testing"
	"time"

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
