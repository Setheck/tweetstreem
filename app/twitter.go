package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/gomodule/oauth1/oauth"
)

var (
	TwitterCredentialRequestURI = "https://api.twitter.com/oauth/request_token"
	TwitterTokenRequestURI      = "https://api.twitter.com/oauth/access_token"
	TwitterAuthorizeURI         = "https://api.twitter.com/oauth/authorize"

	HomeTimelineURI = "https://api.twitter.com/1.1/statuses/home_timeline.json"

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

var NilCompleter = func(document prompt.Document) []prompt.Suggest {
	return nil
}

type TwitterConfiguration struct {
	PollTime   time.Duration `json:"pollTime"`
	UserToken  string        `json:"userToken"`
	UserSecret string        `json:"userSecret"`
}

func (t *TwitterConfiguration) validate() {
	if t.PollTime < time.Minute*2 {
		t.PollTime = time.Minute * 2
	}
}

type Twitter struct {
	configuration *TwitterConfiguration
	Credentials   *oauth.Credentials

	lastTweet   *Tweet
	wg          sync.WaitGroup
	ctx         context.Context
	done        context.CancelFunc
	oauthClient oauth.Client
	lock        sync.Mutex
	debug       bool
}

func NewTwitter(conf *TwitterConfiguration) *Twitter {
	if conf != nil {
		conf.validate()
	}
	ctx, done := context.WithCancel(context.Background())
	return &Twitter{
		configuration: conf,
		ctx:           ctx,
		done:          done,
		oauthClient: oauth.Client{
			TemporaryCredentialRequestURI: TwitterCredentialRequestURI,
			TokenRequestURI:               TwitterTokenRequestURI,
			ResourceOwnerAuthorizationURI: TwitterAuthorizeURI,
			Credentials:                   oauth.Credentials{Token: AppToken, Secret: AppSecret},
		},
	}
}

func (t *Twitter) Init() error {
	// TODO: validate instead of check for empty
	if t.configuration.UserToken == "" || t.configuration.UserSecret == "" {
		return t.Authorize()
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
	OpenBrowser(u)

	code := prompt.Input("Enter Pin: ", NilCompleter)

	credentials, _, err := t.oauthClient.RequestToken(nil, tempCred, code)
	if err != nil {
		return err
	}

	t.configuration.UserToken = credentials.Token
	t.configuration.UserSecret = credentials.Secret
	return nil
}

type GetConf struct {
	count           int
	sinceId         string
	includeEntities bool
}

func (g *GetConf) ToForm() url.Values {
	form := url.Values{}
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

func (t *Twitter) HomeTimeline(conf GetConf) ([]*Tweet, error) {
	rawTweets, err := t.oaGet(HomeTimelineURI, conf)
	if err != nil {
		return nil, err
	}

	var twErr TwError
	_ = json.Unmarshal(rawTweets, &twErr) // We dont' really care if this fails
	if len(twErr.Errors) > 0 {
		return nil, fmt.Errorf(twErr.String())
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

func (t *Twitter) oaGet(u string, conf GetConf) ([]byte, error) {
	cred := &oauth.Credentials{Token: t.configuration.UserToken, Secret: t.configuration.UserSecret}
	resp, err := t.oauthClient.Get(nil, cred, u, conf.ToForm())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
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
			case <-time.After(t.configuration.PollTime):
				if t.debug {
					fmt.Println("Poll happened")
				}
				cfg := GetConf{includeEntities: true}
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

type Entities struct {
}

type Enrichment struct {
}

type User struct {
	Id                   int64        `json:"id"`
	IdStr                string       `json:"id_str"`
	Name                 string       `json:"name"`
	ScreenName           string       `json:"screen_name"`
	Location             *string      `json:"location"`
	Derived              []Enrichment `json:"derived"`
	Url                  *string      `json:"url"`
	Description          *string      `json:"description"`
	Protected            bool         `json:"protected"`
	Verified             bool         `json:"verified"`
	FollowersCount       int          `json:"followers_count"`
	FriendsCount         int          `json:"friends_count"`
	ListedCount          int          `json:"listed_count"`
	FavouritesCount      int          `json:"favourites_count"`
	StatusesCount        int          `json:"statuses_count"`
	CreatedAt            string       `json:"created_at"`
	ProfileBannerUrl     string       `json:"profile_banner_url"`
	ProfileImageUrlHttps string       `json:"profile_image_url_https"`
	DefaultProfile       bool         `json:"default_profile"`
	DefaultProfileImage  bool         `json:"default_profile_image"`
	WithheldInCountries  []string     `json:"withheld_in_countries"`
	WithheldScope        []string     `json:"withheld_scope"`
}

type ReTweetedStatus struct {
}

type Coordinates struct {
}

type Place struct {
}

type Rule struct {
}

type Tweet struct {
	CreatedAt            string       `json:"created_at"`
	ID                   int64        `json:"id"`
	IDStr                string       `json:"id_str"`
	Text                 string       `json:"text"`
	Source               string       `json:"source"`
	Truncated            bool         `json:"truncated"`
	InReplyToStatusId    *int64       `json:"in_reply_to_status_id"`
	InReplyToStatusIdStr *string      `json:"in_reply_to_status_id_str"`
	InReplyToUserId      *int64       `json:"in_reply_to_user_id"`
	InReplyToUserIdStr   *string      `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  *string      `json:"in_reply_to_screen_name"`
	User                 User         `json:"user"`
	Coordinates          *Coordinates `json:"coordinates"`
	Place                *Place       `json:"place"`
	QuotedStatusId       int64        `json:"quoted_status_id"`
	QuotedStatusIdStr    string       `json:"quoted_status_id_str"`
	IsQuoteStatus        bool         `json:"is_quote_status"`
	QuotedStatus         *Tweet       `json:"quoted_status"`
	ReTweetedStatus      *Tweet       `json:"retweeted_status"`
	QuoteCount           *int         `json:"quote_count"`
	ReplyCount           int          `json:"reply_count"`
	ReTweetCount         int          `json:"retweet_count"`
	FavoriteCount        int          `json:"favorite_count"`
	Entities             Entities     `json:"entities"`
	ExtendedEntities     Entities     `json:"extended_entities"`
	Favorited            *bool        `json:"favorited"`
	ReTweeted            bool         `json:"retweeted"`
	PossiblySensitive    *bool        `json:"possibly_sensitive"`
	FilterLevel          string       `json:"filter_level"`
	Lang                 *string      `json:"lang"`
	MatchingRules        []Rule       `json:"matching_rules"`

	// TODO: Additional Attributes
	// https://developer.twitter.com/en/docs/tweets/data-dictionary/overview/tweet-object
}

func (t *Tweet) UsrString() string {
	tstr := t.CreatedAt
	tm, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", t.CreatedAt)
	if err == nil {
		since := time.Since(tm)
		if since < time.Hour*24 {
			tstr = since.Truncate(time.Second).String()
		} else {
			tstr = tm.Format("01/02/2006 15:04:05")
		}
	}

	str := fmt.Sprint(Cyan, t.User.Name, Reset, " ")
	str += fmt.Sprintf("%s@%s%s ", Green, t.User.ScreenName, Reset)
	str += fmt.Sprint(Purple, tstr, " ago", Reset)
	return str
}

func (t *Tweet) StatusString() string {
	agent := ExtractAnchorText(t.Source)
	return fmt.Sprint(
		Cyan, "rt:", t.ReTweetCount, Reset, " ",
		Red, "♥:", t.FavoriteCount, Reset,
		" via ", Blue, agent, Reset)
}

func ExtractAnchorText(anchor string) string {
	anchorTextFind := regexp.MustCompile(`>(.+)<`)
	found := anchorTextFind.FindStringSubmatch(anchor)
	if len(found) > 0 {
		return found[1]
	}
	return ""
}

func (t *Tweet) String() string {
	return html.UnescapeString(t.Text)
}
