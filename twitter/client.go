package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Setheck/tweetstreem/auth"
	"github.com/Setheck/tweetstreem/util"
)

var (
	CredentialRequestURI = "https://api.twitter.com/oauth/request_token"
	TokenRequestURI      = "https://api.twitter.com/oauth/access_token"
	AuthorizeURI         = "https://api.twitter.com/oauth/authorize"

	AccountSettingsURI  = "https://api.twitter.com/1.1/account/settings.json"
	UserTimelineURI     = "https://api.twitter.com/1.1/statuses/user_timeline.json"
	HomeTimelineURI     = "https://api.twitter.com/1.1/statuses/home_timeline.json"
	StatusesUpdateURI   = "https://api.twitter.com/1.1/statuses/update.json"
	FavoritesCreateURI  = "https://api.twitter.com/1.1/favorites/create.json"
	FavoritesDestroyURI = "https://api.twitter.com/1.1/favorites/destroy.json"
	FollowersListURI    = "https://api.twitter.com/1.1/followers/list.json"

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

type Client interface {
	Configuration() Configuration
	Authorize() error
	UpdateStatus(status string, conf url.Values) (*Tweet, error)
	ReTweet(tw *Tweet, conf url.Values) error
	UnReTweet(tw *Tweet, conf url.Values) error
	Like(tw *Tweet, conf url.Values) error
	UnLike(tw *Tweet, conf url.Values) error
	HomeTimeline(conf url.Values) ([]*Tweet, error)
	UserTimeline(conf url.Values) ([]*Tweet, error)
	SetPollerPaused(b bool)
	StartPoller(tweetCh chan<- []*Tweet)
	ScreenName() string
	Shutdown()
}

type Configuration struct {
	PollTime   string `json:"pollTime"`
	UserToken  string `json:"userToken"`
	UserSecret string `json:"userSecret"`
}

func (t *Configuration) PollTimeDuration() time.Duration {
	dur, err := time.ParseDuration(t.PollTime)
	if err != nil {
		t.PollTime = "2m" // default to 2 min
		return time.Minute * 2
	}
	return dur
}

var _ Client = &DefaultClient{}

type DefaultClient struct {
	configuration   *Configuration
	accountSettings *AccountSettings
	pollerPaused    bool
	lastTweet       *Tweet
	wg              sync.WaitGroup
	ctx             context.Context
	done            context.CancelFunc
	oauthFacade     auth.OauthFacade
	lock            sync.Mutex
	debug           bool
}

func NewDefaultClient(conf Configuration) *DefaultClient {
	ctx, done := context.WithCancel(context.Background())
	oaconf := auth.OauthConfig{
		TemporaryCredentialRequestURI: CredentialRequestURI,
		TokenRequestURI:               TokenRequestURI,
		ResourceOwnerAuthorizationURI: AuthorizeURI,
		AppToken:                      AppToken,
		AppSecret:                     AppSecret,
		UserAgent:                     "~TweetStreem~",
		Token:                         conf.UserToken,
		Secret:                        conf.UserSecret,
	}
	return &DefaultClient{
		configuration: &conf,
		ctx:           ctx,
		done:          done,
		oauthFacade:   auth.NewDefaultOaFacade(oaconf),
	}
}

var openBrowser = util.OpenBrowser

func (t *DefaultClient) Configuration() Configuration {
	if t.configuration != nil {
		return *t.configuration
	}
	return Configuration{}
}

// replace fmt.Print here because it breaks parsing of test status output
// ref: https://youtrack.jetbrains.com/issue/GO-7855 & https://github.com/golang/go/issues/23036
var fmtPrint = fmt.Print

func (t *DefaultClient) Authorize() error {
	if t.configuration.UserToken != "" && t.configuration.UserSecret != "" {
		if err := t.updateAccountSettings(); err == nil {
			return nil
		}
	}
	tempCred, err := t.oauthFacade.RequestTemporaryCredentials(nil, "oob", nil)
	if err != nil {
		return err
	}

	u := t.oauthFacade.AuthorizationURL(tempCred, nil)
	if err := openBrowser(u); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	fmtPrint("Enter Pin: ")
	code := util.SingleWordInput()

	credentials, _, err := t.oauthFacade.RequestToken(nil, tempCred, code)
	if err != nil {
		return err
	}

	t.configuration.UserToken = credentials.Token
	t.oauthFacade.SetToken(credentials.Token)
	t.configuration.UserSecret = credentials.Secret
	t.oauthFacade.SetSecret(credentials.Secret)
	if err := t.updateAccountSettings(); err != nil {
		return fmt.Errorf("failed to authorize, couldn't get account settings: %w", err)
	}
	return nil
}

func NewUrlValues() url.Values {
	orc := make(url.Values)

	// defaults
	orc.Set("tweet_mode", "extended") // Default to extended for full tweet text

	return orc
}

type TwError struct {
	Code    int
	Message string
}

type TwErrors struct {
	Errors []TwError
}

func (twe TwErrors) String() string {
	outstr := ""
	for _, e := range twe.Errors {
		outstr += fmt.Sprintf("%d - %s ", e.Code, e.Message)
	}
	return outstr
}

func (t *DefaultClient) UpdateStatus(status string, conf url.Values) (*Tweet, error) {
	conf.Set("status", status)
	data, err := t.oauthFacade.OaRequest(http.MethodPost, StatusesUpdateURI, conf)
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

func (t *DefaultClient) ReTweet(tw *Tweet, conf url.Values) error {
	data, err := t.oauthFacade.OaRequest(http.MethodPost, fmt.Sprintf(StatusesRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *DefaultClient) UnReTweet(tw *Tweet, conf url.Values) error {
	data, err := t.oauthFacade.OaRequest(http.MethodPost, fmt.Sprintf(StatusesUnRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *DefaultClient) Like(tw *Tweet, conf url.Values) error {
	conf.Set("id", tw.IDStr)
	data, err := t.oauthFacade.OaRequest(http.MethodPost, FavoritesCreateURI, conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *DefaultClient) UnLike(tw *Tweet, conf url.Values) error {
	conf.Set("id", tw.IDStr)
	data, err := t.oauthFacade.OaRequest(http.MethodPost, FavoritesDestroyURI, conf)
	if err != nil {
		return err
	}
	if err := t.unmarshalError(data); err != nil {
		return err
	}
	return nil
}

type FollowerList struct {
	Users             []User `json:"users"`
	NextCursor        uint64 `json:"next_cursor"`
	NextCursorStr     string `json:"next_cursor_str"`
	PreviousCursor    uint64 `json:"previous_cursor"`
	PreviousCursorStr string `json:"previous_cursor_str"`
}

func (t *DefaultClient) ListFollowers(conf url.Values) ([]User, error) {
	data, err := t.oauthFacade.OaRequest(http.MethodGet, FollowersListURI, conf)
	if err != nil {
		return nil, err
	}
	if err := t.unmarshalError(data); err != nil {
		return nil, err
	}
	fl := &FollowerList{}
	if err := json.Unmarshal(data, fl); err != nil {
		return nil, err
	}
	return fl.Users, nil
}

func (t *DefaultClient) unmarshalError(data []byte) error {
	var errList TwErrors
	_ = json.Unmarshal(data, &errList) // We dont' really care if this fails
	if len(errList.Errors) > 0 {
		return fmt.Errorf(errList.String())
	}
	return nil
}

func (t *DefaultClient) ScreenName() string {
	if t.accountSettings == nil {
		return ""
	}
	return t.accountSettings.ScreenName
}

func (t *DefaultClient) updateAccountSettings() error {
	raw, err := t.oauthFacade.OaRequest(http.MethodGet, AccountSettingsURI, url.Values{})
	if err != nil {
		return err
	}

	if err := t.unmarshalError(raw); err != nil {
		return err
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	return json.Unmarshal(raw, &t.accountSettings)
}

func (t *DefaultClient) HomeTimeline(conf url.Values) ([]*Tweet, error) {
	return t.getTimeline(HomeTimelineURI, conf)
}

func (t *DefaultClient) UserTimeline(conf url.Values) ([]*Tweet, error) {
	return t.getTimeline(UserTimelineURI, conf)
}

func (t *DefaultClient) getTimeline(timelineUri string, conf url.Values) ([]*Tweet, error) {
	rawTweets, err := t.oauthFacade.OaRequest(http.MethodGet, timelineUri, conf)
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

func (t *DefaultClient) SetPollerPaused(b bool) {
	t.pollerPaused = b
}

// StartPoller will periodically poll and add resulting tweets to the given tweet channel
// When the poller is done (twitter context cancelled) it will close the channel
func (t *DefaultClient) StartPoller(tweetCh chan<- []*Tweet) {
	if t.debug {
		fmt.Println("Poller Started")
	}
	t.wg.Add(1)
	go func(resultCh chan<- []*Tweet) {
		defer func() {
			close(resultCh)
			t.wg.Done()
		}()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-time.After(t.configuration.PollTimeDuration()):
				if t.debug {
					fmt.Println("Poll happened")
				}
				if t.pollerPaused {
					continue
				}
				cfg := NewUrlValues()
				cfg.Set("include_entities", "true")
				if t.lastTweet != nil {
					t.lock.Lock()
					cfg.Set("since_id", t.lastTweet.IDStr)
					t.lock.Unlock()
				}
				tweets, err := t.HomeTimeline(cfg)
				if err != nil {
					fmt.Println("Poll Failure:", err)
				} else {
					resultCh <- tweets
				}
			}
		}
	}(tweetCh)
}

func (t *DefaultClient) Shutdown() {
	t.done()
	t.wg.Wait()
}