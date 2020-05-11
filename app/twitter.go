package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	TwitterCredentialRequestURI = "https://api.twitter.com/oauth/request_token"
	TwitterTokenRequestURI      = "https://api.twitter.com/oauth/access_token"
	TwitterAuthorizeURI         = "https://api.twitter.com/oauth/authorize"

	AccountSettingsURI  = "https://api.twitter.com/1.1/account/settings.json"
	UserTimelineURI     = "https://api.twitter.com/1.1/statuses/user_timeline.json"
	HomeTimelineURI     = "https://api.twitter.com/1.1/statuses/home_timeline.json"
	StatusesUpdateURI   = "https://api.twitter.com/1.1/statuses/update.json"
	FavoritesCreateURI  = "https://api.twitter.com/1.1/favorites/create.json"
	FavoritesDestroyURI = "https://api.twitter.com/1.1/favorites/destroy.json"

	StatusesRetweetURITemplate   = "https://api.twitter.com/1.1/statuses/retweet/%s.json"
	StatusesUnRetweetURITemplate = "https://api.twitter.com/1.1/statuses/unretweet/%s.json"

	TweetLinkUriTemplate = "https://twitter.com/%s/status/%s"

	AppToken  = ""
	AppSecret = ""
)

func init() {
	// Since the plan is to stamp the AppToken and AppSecret on deploy,
	// this allows us to pull it out of the ENV for dev/testing.
	if AppToken == "" || AppSecret == "" {
		AppToken = os.Getenv("APP_TOKEN")
		AppSecret = os.Getenv("APP_SECRET")
		fmt.Println("loaded AppToken and AppSecret from environment")
	}
}

type TwitterClient interface {
	Init() error
	Configuration() TwitterConfiguration
	Authorize() error
	UpdateStatus(status string, conf OaRequestConf) (*Tweet, error)
	ReTweet(tw *Tweet, conf OaRequestConf) error
	UnReTweet(tw *Tweet, conf OaRequestConf) error
	Like(tw *Tweet, conf OaRequestConf) error
	UnLike(tw *Tweet, conf OaRequestConf) error
	HomeTimeline(conf OaRequestConf) ([]*Tweet, error)
	UserTimeline(conf OaRequestConf) ([]*Tweet, error)
	TogglePollerPaused(b bool)
	ScreenName() string
	Shutdown() error
}

type TwitterConfiguration struct {
	PollTime   string `json:"pollTime"`
	UserToken  string `json:"userToken"`
	UserSecret string `json:"userSecret"`
}

func (t *TwitterConfiguration) PollTimeDuration() time.Duration {
	dur, err := time.ParseDuration(t.PollTime)
	if err != nil {
		t.PollTime = "2m" // default to 2 min
		return time.Minute * 2
	}
	return dur
}

var _ TwitterClient = &Twitter{}

type Twitter struct {
	configuration   *TwitterConfiguration
	accountSettings *AccountSettings
	pollerPaused    bool
	lastTweet       *Tweet
	wg              sync.WaitGroup
	ctx             context.Context
	done            context.CancelFunc
	oauthClient     OauthFacade
	lock            sync.Mutex
	debug           bool
}

func NewTwitter(conf *TwitterConfiguration) *Twitter {
	ctx, done := context.WithCancel(context.Background())
	oaconf := OauthConfig{
		TemporaryCredentialRequestURI: TwitterCredentialRequestURI,
		TokenRequestURI:               TwitterTokenRequestURI,
		ResourceOwnerAuthorizationURI: TwitterAuthorizeURI,
		AppToken:                      AppToken,
		AppSecret:                     AppSecret,
		UserAgent:                     "~TweetStreem~",
		Token:                         conf.UserToken,
		Secret:                        conf.UserSecret,
	}
	return &Twitter{
		configuration: conf,
		ctx:           ctx,
		done:          done,
		oauthClient:   NewDefaultOaFacade(oaconf),
	}
}

func (t *Twitter) Init() error {
	// TODO: validate instead of check for empty
	if t.configuration.UserToken == "" || t.configuration.UserSecret == "" {
		if err := t.Authorize(); err != nil {
			return err
		}
	}
	if err := t.updateAccountSettings(); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) Configuration() TwitterConfiguration {
	return *t.configuration
}

func (t *Twitter) Authorize() error {
	tempCred, err := t.oauthClient.RequestTemporaryCredentials(nil, "oob", nil)
	if err != nil {
		return err
	}

	u := t.oauthClient.AuthorizationURL(tempCred, nil)
	if err := OpenBrowser(u); err != nil {
		fmt.Println(err)
	}

	fmt.Print("Enter Pin: ")
	code := SingleWordInput()

	credentials, _, err := t.oauthClient.RequestToken(nil, tempCred, code)
	if err != nil {
		return err
	}

	t.configuration.UserToken = credentials.Token
	t.oauthClient.SetToken(credentials.Token)
	t.configuration.UserSecret = credentials.Secret
	t.oauthClient.SetSecret(credentials.Secret)
	return nil
}

type OaRequestConf struct {
	id                string
	status            string
	screenName        string
	count             int
	sinceId           string
	includeEntities   bool
	tweetMode         string
	InReplyToStatusId string
}

func (g *OaRequestConf) ToForm() url.Values {
	form := url.Values{}
	if len(g.id) > 0 {
		form.Set("id", g.id)
	}
	if len(g.status) > 0 {
		form.Set("status", g.status)
	}
	if len(g.InReplyToStatusId) > 0 {
		form.Set("in_reply_to_status_id", g.InReplyToStatusId)
	}
	if len(g.screenName) > 0 {
		form.Set("screen_name", g.screenName)
	}
	if g.count > 0 {
		cnt := strconv.Itoa(g.count)
		form.Set("count", cnt)
	}
	if len(g.sinceId) > 0 {
		form.Set("since_id", g.sinceId)
	}
	if g.includeEntities {
		form.Set("include_entities", "true")
	}
	mode := "extended" // Default to extended for full tweet text
	if len(g.tweetMode) > 0 {
		mode = g.tweetMode
	}
	form.Set("tweet_mode", mode)

	return form
}

type TwError struct {
	Errors []struct {
		Code    int
		Message string
	}
}

func (twe TwError) String() string {
	outstr := ""
	for _, e := range twe.Errors {
		outstr += fmt.Sprintf("%d - %s ", e.Code, e.Message)
	}
	return outstr
}

func (t *Twitter) UpdateStatus(status string, conf OaRequestConf) (*Tweet, error) {
	conf.status = status
	data, err := t.oauthClient.OaRequest(http.MethodPost, StatusesUpdateURI, conf)
	if err != nil {
		return nil, err
	}
	if err := t.unmarshalError(data); err != nil {
		return nil, err
	}
	tw := new(Tweet)
	if err := json.Unmarshal(data, &tw); err != nil {
		return nil, err
	}
	return tw, nil
}

func (t *Twitter) ReTweet(tw *Tweet, conf OaRequestConf) error {
	data, err := t.oauthClient.OaRequest(http.MethodPost, fmt.Sprintf(StatusesRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) UnReTweet(tw *Tweet, conf OaRequestConf) error {
	data, err := t.oauthClient.OaRequest(http.MethodPost, fmt.Sprintf(StatusesUnRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) Like(tw *Tweet, conf OaRequestConf) error {
	conf.id = tw.IDStr
	data, err := t.oauthClient.OaRequest(http.MethodPost, FavoritesCreateURI, conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) UnLike(tw *Tweet, conf OaRequestConf) error {
	conf.id = tw.IDStr
	data, err := t.oauthClient.OaRequest(http.MethodPost, FavoritesDestroyURI, conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) unmarshalError(data []byte) error {
	var twErr TwError
	_ = json.Unmarshal(data, &twErr) // We dont' really care if this fails
	if len(twErr.Errors) > 0 {
		return fmt.Errorf(twErr.String())
	}
	return nil
}

func (t *Twitter) ScreenName() string {
	return t.accountSettings.ScreenName
}

func (t *Twitter) updateAccountSettings() error {
	raw, err := t.oauthClient.OaRequest(http.MethodGet, AccountSettingsURI, OaRequestConf{})
	if err != nil {
		return err
	}

	if err := t.unmarshalError(raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, &t.accountSettings)
}

func (t *Twitter) HomeTimeline(conf OaRequestConf) ([]*Tweet, error) {
	rawTweets, err := t.oauthClient.OaRequest(http.MethodGet, HomeTimelineURI, conf)
	if err != nil {
		return nil, err
	}

	if err := t.unmarshalError(rawTweets); err != nil {
		return nil, err
	}
	var timeLine []*Tweet
	if err := json.Unmarshal(rawTweets, &timeLine); err != nil {
		return nil, err
	}
	if len(timeLine) > 0 {
		t.lock.Lock()
		t.lastTweet = timeLine[0]
		t.lock.Unlock()
	}
	return timeLine, nil
}

func (t *Twitter) UserTimeline(conf OaRequestConf) ([]*Tweet, error) {
	rawTweets, err := t.oauthClient.OaRequest(http.MethodGet, UserTimelineURI, conf)
	if err != nil {
		return nil, err
	}

	if err := t.unmarshalError(rawTweets); err != nil {
		return nil, err
	}
	var timeLine []*Tweet
	if err := json.Unmarshal(rawTweets, &timeLine); err != nil {
		return nil, err
	}
	if len(timeLine) > 0 {
		t.lock.Lock()
		t.lastTweet = timeLine[0]
		t.lock.Unlock()
	}
	return timeLine, nil
}

func (t *Twitter) TogglePollerPaused(b bool) {
	t.pollerPaused = b
}

func (t *Twitter) startPoller() chan []*Tweet {
	if t.debug {
		fmt.Println("Poller Started")
	}
	tweetCh := make(chan []*Tweet, 0)
	go func() {
		t.wg.Add(1)
		defer t.wg.Done()
		for {
			select {
			case <-t.ctx.Done():
				close(tweetCh)
				return
			case <-time.After(t.configuration.PollTimeDuration()):
				if t.debug {
					fmt.Println("Poll happened")
				}
				if t.pollerPaused {
					continue
				}
				cfg := OaRequestConf{includeEntities: true}
				if t.lastTweet != nil {
					t.lock.Lock()
					cfg.sinceId = t.lastTweet.IDStr
					t.lock.Unlock()
				}
				tweets, err := t.HomeTimeline(cfg)
				if err != nil {
					fmt.Println("Poll Failure:", err)
				} else {
					// get new tweets
					tweetCh <- tweets
				}
			}
		}
	}()

	return tweetCh
}

func (t *Twitter) Shutdown() error {
	t.done()
	t.wg.Wait()
	return nil
}
