package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/gomodule/oauth1/oauth"
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
	if conf != nil {
		conf.validate()
	}
	ctx, done := context.WithCancel(context.Background())
	oaconf := OauthConfig{
		TemporaryCredentialRequestURI: TwitterCredentialRequestURI,
		TokenRequestURI:               TwitterTokenRequestURI,
		ResourceOwnerAuthorizationURI: TwitterAuthorizeURI,
		Credentials:                   oauth.Credentials{Token: AppToken, Secret: AppSecret},
	}
	return &Twitter{
		configuration: conf,
		ctx:           ctx,
		done:          done,
		oauthClient:   NewOaClient(oaconf),
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

	code := prompt.Input("Enter Pin: ", NilCompleter)

	credentials, _, err := t.oauthClient.RequestToken(nil, tempCred, code)
	if err != nil {
		return err
	}

	t.configuration.UserToken = credentials.Token
	t.configuration.UserSecret = credentials.Secret
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
	data, err := t.oaRequest(http.MethodPost, StatusesUpdateURI, conf)
	if err != nil {
		return nil, err
	}
	if err := t.UnmarshalError(data); err != nil {
		return nil, err
	}
	tw := new(Tweet)
	if err := json.Unmarshal(data, &tw); err != nil {
		return nil, err
	}
	return tw, nil
}

func (t *Twitter) ReTweet(tw *Tweet, conf OaRequestConf) error {
	data, err := t.oaRequest(http.MethodPost, fmt.Sprintf(StatusesRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.UnmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) UnReTweet(tw *Tweet, conf OaRequestConf) error {
	data, err := t.oaRequest(http.MethodPost, fmt.Sprintf(StatusesUnRetweetURITemplate, tw.IDStr), conf)
	if err != nil {
		return err
	}
	if err := t.UnmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) Like(tw *Tweet, conf OaRequestConf) error {
	conf.id = tw.IDStr
	data, err := t.oaRequest(http.MethodPost, FavoritesCreateURI, conf)
	if err != nil {
		return err
	}
	if err := t.UnmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) UnLike(tw *Tweet, conf OaRequestConf) error {
	conf.id = tw.IDStr
	data, err := t.oaRequest(http.MethodPost, FavoritesDestroyURI, conf)
	if err != nil {
		return err
	}
	if err := t.UnmarshalError(data); err != nil {
		return err
	}
	return nil
}

func (t *Twitter) UnmarshalError(data []byte) error {
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
	raw, err := t.oaRequest(http.MethodGet, AccountSettingsURI, OaRequestConf{})
	if err != nil {
		return err
	}

	if err := t.UnmarshalError(raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, &t.accountSettings)
}

func (t *Twitter) HomeTimeline(conf OaRequestConf) ([]*Tweet, error) {
	rawTweets, err := t.oaRequest(http.MethodGet, HomeTimelineURI, conf)
	if err != nil {
		return nil, err
	}

	if err := t.UnmarshalError(rawTweets); err != nil {
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
	rawTweets, err := t.oaRequest(http.MethodGet, UserTimelineURI, conf)
	if err != nil {
		return nil, err
	}

	if err := t.UnmarshalError(rawTweets); err != nil {
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

func (t *Twitter) oaRequest(method, u string, conf OaRequestConf) ([]byte, error) {
	cred := &oauth.Credentials{Token: t.configuration.UserToken, Secret: t.configuration.UserSecret}
	var resp *http.Response
	var err error
	formData := conf.ToForm()
	formData.Set("User-Agent", "tweetstream-"+Version)
	switch strings.ToUpper(method) {
	case http.MethodPost:
		resp, err = t.oauthClient.Post(nil, cred, u, formData)
	case http.MethodGet:
		resp, err = t.oauthClient.Get(nil, cred, u, formData)
	}
	if err != nil {
		return nil, err
	}
	if resp != nil {
		if resp.StatusCode != http.StatusOK { // TODO: only non 200 ?
			return nil, fmt.Errorf("failed: %d - %s", resp.StatusCode, resp.Status)
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("unsupported method")
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
			case <-time.After(t.configuration.PollTime):
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

type HashTag struct {
	Indices []int  `json:"indices"`
	Text    string `json:"text"`
}

type UserMention struct {
	Name       string `json:"name"`
	Indices    []int  `json:"indices"`
	ScreenName string `json:"screen_name"`
	Id         int64  `json:"id"`
	IdStr      string `json:"id_str"`
}

type Symbol struct {
	// TODO
}

type Url struct {
	DisplayUrl  string `json:"display_url"`
	ExpandedUrl string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
	Url         string `json:"url"`
}

type Media struct {
	DisplayUrl    string `json:"display_url"`
	ExpandedUrl   string `json:"expanded_url"`
	Id            int64  `json:"id"`
	IdStr         string `json:"id_str"`
	Indices       []int  `json:"indices"`
	MediaUrl      string `json:"media_url"`
	MediaUrlHttps string `json:"media_url_https"`
	//Sizes - TODO
	Type string `json:"photo"`
	Url  string `json:"url"`
}

type Entities struct {
	HashTags    []HashTag     `json:"hashtags"`
	Urls        []Url         `json:"urls"`
	UserMention []UserMention `json:"user_mentions"`
	Symbol      []Symbol      `json:"symbols"`
	Media       []Media       `json:"media"`
}

type Enrichment struct {
	// TODO:
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
	// TODO:
}

type Coordinates struct {
	// TODO:
}

type Place struct {
	// TODO:
}

type Rule struct {
	// TODO:
}

type Tweet struct {
	CreatedAt            string       `json:"created_at"`
	ID                   int64        `json:"id"`
	IDStr                string       `json:"id_str"`
	Text                 string       `json:"text"`
	FullText             string       `json:"full_text"`
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

func (t *Tweet) HtmlLink() string {
	return fmt.Sprintf(TweetLinkUriTemplate, t.User.ScreenName, t.IDStr)
}

func (t *Tweet) Links() []string {
	ulist := make([]string, 0)
	for _, u := range t.Entities.Urls {
		ulist = append(ulist, u.ExpandedUrl)
	}
	for _, u := range t.Entities.Media {
		ulist = append(ulist, u.MediaUrl)
	}
	return ulist
}

type TweetTemplateOutput struct {
	UserName          string
	ScreenName        string
	RelativeTweetTime string
	ReTweetCount      string
	FavoriteCount     string
	App               string
	TweetText         string
}

type OutputConfig struct {
	MentionHighlightColor string
	HashtagHighlightColor string
}

func (t *Tweet) TemplateOutput(config OutputConfig) TweetTemplateOutput {
	return TweetTemplateOutput{
		UserName:          t.User.Name,
		ScreenName:        t.User.ScreenName,
		RelativeTweetTime: t.RelativeTweetTime(),
		ReTweetCount:      strconv.Itoa(t.ReTweetCount),
		FavoriteCount:     strconv.Itoa(t.FavoriteCount),
		App:               ExtractAnchorText(t.Source),
		TweetText:         t.TweetText(config, true),
	}
}

func (t *Tweet) RelativeTweetTime() string {
	tstr := t.CreatedAt
	tm, err := time.Parse("Mon Jan 2 15:04:05 -0700 2006", t.CreatedAt)
	if err == nil {
		since := time.Since(tm)
		if since < time.Hour*24 {
			tstr = since.Truncate(time.Second).String() + " ago"
		} else {
			tstr = tm.Format("01/02/2006 15:04:05")
		}
	}
	return tstr
}

func (t *Tweet) highlightEntities(config OutputConfig) HighlightEntityList {
	var hlents HighlightEntityList
	for _, ht := range t.Entities.HashTags {
		start, end := ht.Indices[0], ht.Indices[1]
		hlents = append(hlents, HighlightEntity{
			startIdx: start,
			endIdx:   end,
			color:    config.HashtagHighlightColor})
	}
	for _, um := range t.Entities.UserMention {
		start, end := um.Indices[0], um.Indices[1]
		hlents = append(hlents, HighlightEntity{
			startIdx: start,
			endIdx:   end,
			color:    config.MentionHighlightColor})
	}
	return hlents
}

func (t *Tweet) TweetText(config OutputConfig, highlight bool) string {
	text := ""
	if t.ReTweetedStatus != nil {
		text = t.ReTweetedStatus.TweetText(config, highlight)
		screenName := Colors.Colorize(config.MentionHighlightColor, "@"+t.ReTweetedStatus.User.ScreenName)
		return fmt.Sprintf("RT %s: %s", screenName, text)
	}
	if len(t.FullText) > 0 {
		text = t.FullText
	} else {
		text = t.Text
	}
	if highlight {
		list := t.highlightEntities(config)
		text = highlightEntries(text, list)
	}
	return html.UnescapeString(text)
}

type HighlightEntity struct {
	startIdx int
	endIdx   int
	color    string
}

type HighlightEntityList []HighlightEntity

func (l HighlightEntityList) Len() int {
	return len(l)
}
func (l HighlightEntityList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l HighlightEntityList) Less(i, j int) bool {
	return l[i].startIdx < l[j].startIdx
}

func highlightEntries(text string, hlist HighlightEntityList) string {
	sort.Sort(hlist)
	rtext := []rune(text)
	resultText := ""
	position := 0
	for _, entry := range hlist {
		resultText += string(rtext[position:entry.startIdx])
		resultText += Colors.Colorize(entry.color, string(rtext[entry.startIdx:entry.endIdx]))
		position = entry.endIdx
	}
	resultText += string(rtext[position:])
	return resultText
}

type SleepTime struct {
	Enabled   bool   `json:"enabled"`
	EndTime   *int64 `json:"end_time"`
	StartTime *int64 `json:"start_time"`
}

type TimeZone struct {
	Name       string `json:"name"`
	TzinfoName string `json:"tzinfo_name"`
	UtcOffset  int64  `json:"utc_offset"`
}

type PlaceType struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

type TrendLocation struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
	ParentId    int64  `json:"parentid"`
	PlaceType   `json:"placeType"`
	Url         string `json:"url"`
	Woeid       int64  `json:"woeid"`
}

type AccountSettings struct {
	AlwaysUseHttps           bool   `json:"always_use_https"`
	DiscoverableByEmail      bool   `json:"discoverable_by_email"`
	GeoEnabled               bool   `json:"geo_enabled"`
	Language                 string `json:"language"`
	Protected                bool   `json:"protected"`
	ScreenName               string `json:"screen_name"`
	ShowAllInlineMedia       bool   `json:"show_all_inline_media"`
	SleepTime                `json:"sleep_time"`
	TimeZone                 `json:"time_zone"`
	TrendLocation            []TrendLocation `json:"trend_location"`
	UseCookiePersonalization bool            `json:"use_cookie_personalization"`
	AllowContributorRequest  string          `json:"allow_contributor_request"`
}
