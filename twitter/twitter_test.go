package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/Setheck/tweetstreem/auth/mocks"
	"github.com/Setheck/tweetstreem/util"
	"github.com/gomodule/oauth1/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultClient_ScreenName(t *testing.T) {
	twitter := &DefaultClient{}
	assert.Equal(t, "", twitter.ScreenName())

	screenName := "some screen name"
	twitter.accountSettings = &AccountSettings{ScreenName: screenName}
	assert.Equal(t, screenName, twitter.ScreenName())
}

func TestDefaultClient_SetPollerPaused(t *testing.T) {
	twitter := &DefaultClient{}
	assert.False(t, twitter.pollerPaused)

	twitter.SetPollerPaused(true)
	assert.True(t, twitter.pollerPaused)

	twitter.SetPollerPaused(false)
	assert.False(t, twitter.pollerPaused)
}

func TestDefaultClient_StartPoller(t *testing.T) {
	// TODO:(smt) complete test
	pollDuration := 3 * time.Millisecond
	tweetOutput := createTwitterResponseData(t, []*Tweet{{IDStr: "12345"}})
	tests := []struct {
		name  string
		polls int
	}{
		//{"", 0},
		{"", 1},
		{"", 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauthFacade := new(mocks.OauthFacade)
			mockOauthFacade.On("OaRequest",
				http.MethodGet,
				HomeTimelineURI,
				mock.AnythingOfType("url.Values")).
				Return(tweetOutput, nil)

			duration := pollDuration * time.Duration(test.polls+1)
			twitter := NewDefaultClient(&Configuration{PollTime: pollDuration.String()})
			twitter.oauthFacade = mockOauthFacade
			tweetCh := make(chan []*Tweet, 10)
			twitter.StartPoller(tweetCh)
			<-time.After(duration)
			assert.Len(t, tweetCh, test.polls)
			twitter.Shutdown()
			mockOauthFacade.AssertNumberOfCalls(t, "OaRequest", test.polls)
		})
	}
}

func TestConfiguration_PollTimeDuration(t *testing.T) {
	tests := []struct {
		name             string
		expectedDuration time.Duration
	}{
		{"", 2 * time.Minute}, // Default
		{"10s", 10 * time.Second},
		{"4m", 4 * time.Minute},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			twcfg := &Configuration{PollTime: test.name}
			assert.Equal(t, test.expectedDuration, twcfg.PollTimeDuration())
		})
	}
}

func TestDefaultClient_Authorize(t *testing.T) {
	// Replace fmtPrint, to prevent test result status parsing failure
	fmtPrint = func(a ...interface{}) (n int, err error) { return 0, nil }

	codeInput := "12345"
	token := "testTokenAsdf123"
	secret := "secretShhShh"

	tests := []struct {
		name           string
		rqTmpCredErr   error
		openBrowserErr error
		rqTokenErr     error
		oaRequestErr   error
	}{
		{"success", nil, nil, nil, nil},
		{"request temp cred failure", assert.AnError, nil, nil, nil},
		{"open browser failure", nil, assert.AnError, nil, nil},
		{"request token failure", nil, nil, assert.AnError, nil},
		{"get account fail", nil, nil, nil, assert.AnError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				mock.AnythingOfType("string"),
				AccountSettingsURI,
				mock.AnythingOfType("url.Values"),
			).Return(createTwitterResponseData(t, &AccountSettings{}), test.oaRequestErr)

			mockOauth.On("RequestTemporaryCredentials",
				mock.AnythingOfType("*http.Client"),
				"oob",
				mock.AnythingOfType("url.Values"),
			).Return(&oauth.Credentials{}, test.rqTmpCredErr)

			mockOauth.On("AuthorizationURL",
				mock.AnythingOfType("*oauth.Credentials"),
				mock.AnythingOfType("url.Values"),
			).Return("someUrl")

			// Prevent browsers from opening
			// This is essentially the openBrowser mock
			openBrowser = func(url string) error { return test.openBrowserErr }

			mockOauth.On("RequestToken",
				mock.AnythingOfType("*http.Client"),
				mock.AnythingOfType("*oauth.Credentials"),
				mock.MatchedBy(func(str string) bool {
					return str == codeInput
				}),
			).Return(&oauth.Credentials{Token: token, Secret: secret}, nil, test.rqTokenErr)

			mockOauth.On("SetToken", mock.MatchedBy(func(str string) bool {
				return str == token
			}))
			mockOauth.On("SetSecret", mock.MatchedBy(func(str string) bool {
				return str == secret
			}))

			util.Stdin = bytes.NewBuffer([]byte(codeInput))

			twitter := NewDefaultClient(&Configuration{})
			twitter.oauthFacade = mockOauth

			err := twitter.Authorize()
			if anyNonNil(t, test.rqTmpCredErr, test.openBrowserErr, test.rqTokenErr, test.oaRequestErr) {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, token, twitter.configuration.UserToken)
				assert.Equal(t, secret, twitter.configuration.UserSecret)
			}
		})
	}
}

func TestDefaultClient_AlreadyAuth(t *testing.T) {
	userToken, userSecret := "user token", "user secret"

	mockOauth := new(mocks.OauthFacade)
	mockOauth.On("OaRequest",
		mock.AnythingOfType("string"),
		AccountSettingsURI,
		mock.AnythingOfType("url.Values"),
	).Return(createTwitterResponseData(t, &AccountSettings{}), nil)

	twitter := NewDefaultClient(&Configuration{
		UserToken:  userToken,
		UserSecret: userSecret})
	twitter.oauthFacade = mockOauth

	err := twitter.Authorize()
	assert.NoError(t, err)
	mockOauth.AssertExpectations(t)
}

func TestNewUrlValues(t *testing.T) {
	uvs := NewUrlValues()

	defaultValues := url.Values{
		"tweet_mode": []string{"extended"},
	}

	assert.Equal(t, uvs, defaultValues)
}

func TestDefaultClient_UpdateStatus(t *testing.T) {
	status := "testing"
	resultTweet := &Tweet{Text: status}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", createTwitterResponseData(t, resultTweet), nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"marshal error", []byte("garbage"), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodPost,
				StatusesUpdateURI,
				mock.MatchedBy(func(uv url.Values) bool {
					return uv.Get("status") == status
				}),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			tweet, err := twitter.UpdateStatus(status, url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, resultTweet, tweet)
			}
		})
	}
}

func TestDefaultClient_ReTweet(t *testing.T) {
	tweet := &Tweet{IDStr: "123"}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", nil, nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodPost,
				mock.MatchedBy(func(str string) bool {
					return str == fmt.Sprintf(StatusesRetweetURITemplate, tweet.IDStr)
				}),
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			err := twitter.ReTweet(tweet, url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultClient_UnReTweet(t *testing.T) {
	tweet := &Tweet{IDStr: "123"}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", nil, nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodPost,
				mock.MatchedBy(func(str string) bool {
					return str == fmt.Sprintf(StatusesUnRetweetURITemplate, tweet.IDStr)
				}),
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			err := twitter.UnReTweet(tweet, url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultClient_Like(t *testing.T) {
	tweet := &Tweet{IDStr: "123"}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", nil, nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodPost,
				FavoritesCreateURI,
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			err := twitter.Like(tweet, url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultClient_UnLike(t *testing.T) {
	tweet := &Tweet{IDStr: "123"}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", nil, nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodPost,
				FavoritesDestroyURI,
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			err := twitter.UnLike(tweet, url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultClient_ListFollowers(t *testing.T) {
	expectedFollowers := []User{{
		Id:         123,
		IdStr:      "123",
		Name:       "TestUser",
		ScreenName: "TestUser",
	}}
	// For the response, we need to add the api cursor wrapper
	followersResponse := createTwitterResponseData(t, &FollowerList{Users: expectedFollowers})

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", followersResponse, nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"marshal error", []byte("garbage"), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodGet,
				FollowersListURI,
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			followers, err := twitter.ListFollowers(url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedFollowers, followers)
			}
		})
	}
}

func TestDefaultClient_UserTimeline(t *testing.T) {
	expectedTweets := []*Tweet{
		{ID: 123, IDStr: "123"},
		{ID: 1243, IDStr: "1243"},
		{ID: 1223, IDStr: "1223"},
		{ID: 11123, IDStr: "11123"},
	}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", createTwitterResponseData(t, expectedTweets), nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"marshal error", []byte("garbage"), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodGet,
				UserTimelineURI,
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			tweets, err := twitter.UserTimeline(url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedTweets, tweets)
				assert.Equal(t, twitter.lastTweet, tweets[0])
			}
		})
	}
}

func TestDefaultClient_HomeTimeline(t *testing.T) {
	expectedTweets := []*Tweet{
		{ID: 123, IDStr: "123"},
		{ID: 1243, IDStr: "1243"},
		{ID: 1223, IDStr: "1223"},
		{ID: 11123, IDStr: "11123"},
	}

	tests := []struct {
		name        string
		tweetData   []byte
		tweetError  error
		expectError bool
	}{
		{"success", createTwitterResponseData(t, expectedTweets), nil, false},
		{"api error", createTwitterErrorData(t), nil, true},
		{"marshal error", []byte("garbage"), nil, true},
		{"request error", nil, assert.AnError, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockOauth := new(mocks.OauthFacade)
			mockOauth.On("OaRequest",
				http.MethodGet,
				HomeTimelineURI,
				mock.AnythingOfType("url.Values"),
			).Return(test.tweetData, test.tweetError)

			twitter := &DefaultClient{}
			twitter.oauthFacade = mockOauth
			tweets, err := twitter.HomeTimeline(url.Values{})
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedTweets, tweets)
				assert.Equal(t, twitter.lastTweet, tweets[0])
			}
		})
	}
}

// Test helper to create a json blob simulating an error coming from the twitter api
func createTwitterErrorData(t *testing.T) []byte {
	t.Helper()
	twErr := TwErrors{
		Errors: []TwError{{
			Code:    0,
			Message: "error",
		}},
	}
	b, err := json.Marshal(twErr)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Test helper to create a json blob simulating a response coming from the twitter api
func createTwitterResponseData(t *testing.T, obj interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func anyNonNil(t *testing.T, objs ...interface{}) bool {
	t.Helper()
	for _, obj := range objs {
		if obj != nil {
			return true
		}
	}
	return false
}
