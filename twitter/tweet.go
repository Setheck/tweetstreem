package twitter

import (
	"fmt"
	"html"
	"strconv"
	"time"

	"github.com/Setheck/tweetstreem/util"
)

const (
	// CreatedAtTimeLayout is the golang time layout for twitter's created time.
	CreatedAtTimeLayout = "Mon Jan 2 15:04:05 -0700 2006"

	// RelativeTweetTimeOutputLayout is the golang time layout that defaults as the tweet time.
	RelativeTweetTimeOutputLayout = "01/02/2006 15:04:05"
)

// HashTag - from the twitter api
type HashTag struct {
	Indices []int  `json:"indices"`
	Text    string `json:"text"`
}

// UserMention - from the twitter api
type UserMention struct {
	Name       string `json:"name"`
	Indices    []int  `json:"indices"`
	ScreenName string `json:"screen_name"`
	ID         int64  `json:"id"`
	IDStr      string `json:"id_str"`
}

// Symbol - from the twitter api
type Symbol struct {
	// TODO
}

// URL - from the twitter api
type URL struct {
	DisplayURL  string `json:"display_url"`
	ExpandedURL string `json:"expanded_url"`
	Indices     []int  `json:"indices"`
	URL         string `json:"url"`
}

// Media - from the twitter api
type Media struct {
	DisplayURL    string `json:"display_url"`
	ExpandedURL   string `json:"expanded_url"`
	ID            int64  `json:"id"`
	IDStr         string `json:"id_str"`
	Indices       []int  `json:"indices"`
	MediaURL      string `json:"media_url"`
	MediaURLHTTPS string `json:"media_url_https"`
	//Sizes - TODO
	Type string `json:"photo"`
	URL  string `json:"url"`
}

// Entities - from the twitter api
type Entities struct {
	HashTags    []HashTag     `json:"hashtags"`
	Urls        []URL         `json:"urls"`
	UserMention []UserMention `json:"user_mentions"`
	Symbol      []Symbol      `json:"symbols"`
	Media       []Media       `json:"media"`
}

// Enrichment - from the twitter api
type Enrichment struct {
	// TODO:
}

// User - from the twitter api
type User struct {
	ID                   int64        `json:"id"`
	IDStr                string       `json:"id_str"`
	Name                 string       `json:"name"`
	ScreenName           string       `json:"screen_name"`
	Location             *string      `json:"location"`
	Derived              []Enrichment `json:"derived"`
	URL                  *string      `json:"url"`
	Description          *string      `json:"description"`
	Protected            bool         `json:"protected"`
	Verified             bool         `json:"verified"`
	FollowersCount       int          `json:"followers_count"`
	FriendsCount         int          `json:"friends_count"`
	ListedCount          int          `json:"listed_count"`
	FavouritesCount      int          `json:"favourites_count"`
	StatusesCount        int          `json:"statuses_count"`
	CreatedAt            string       `json:"created_at"`
	ProfileBannerURL     string       `json:"profile_banner_url"`
	ProfileImageURLHTTPS string       `json:"profile_image_url_https"`
	DefaultProfile       bool         `json:"default_profile"`
	DefaultProfileImage  bool         `json:"default_profile_image"`
	WithheldInCountries  []string     `json:"withheld_in_countries"`
	WithheldScope        []string     `json:"withheld_scope"`
}

// ReTweetedStatus - from the twitter api
type ReTweetedStatus struct {
	// TODO:
}

// Coordinates - from the twitter api
type Coordinates struct {
	// TODO:
}

// Place - from the twitter api
type Place struct {
	// TODO:
}

// Rule - from the twitter api
type Rule struct {
	// TODO:
}

// Tweet - from the twitter api
type Tweet struct {
	CreatedAt            string       `json:"created_at"`
	ID                   int64        `json:"id"`
	IDStr                string       `json:"id_str"`
	Text                 string       `json:"text"`
	FullText             string       `json:"full_text"`
	Source               string       `json:"source"`
	Truncated            bool         `json:"truncated"`
	InReplyToStatusID    *int64       `json:"in_reply_to_status_id"`
	InReplyToStatusIDStr *string      `json:"in_reply_to_status_id_str"`
	InReplyToUserID      *int64       `json:"in_reply_to_user_id"`
	InReplyToUserIDStr   *string      `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  *string      `json:"in_reply_to_screen_name"`
	User                 User         `json:"user"`
	Coordinates          *Coordinates `json:"coordinates"`
	Place                *Place       `json:"place"`
	QuotedStatusID       int64        `json:"quoted_status_id"`
	QuotedStatusIDStr    string       `json:"quoted_status_id_str"`
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

// HTMLLink returns the link to the tweet itself
func (t *Tweet) HTMLLink() string {
	return fmt.Sprintf(TweetLinkUriTemplate, t.User.ScreenName, t.IDStr)
}

// Links returns relevant links from a tweet
func (t *Tweet) Links() []string {
	ulist := make([]string, 0)
	for _, u := range t.Entities.Urls {
		ulist = append(ulist, u.ExpandedURL)
	}
	for _, u := range t.Entities.Media {
		ulist = append(ulist, u.MediaURL)
	}
	return ulist
}

// TweetTemplateOutput is the processed object for use with template execution
type TweetTemplateOutput struct {
	CreatedAt         string
	UserName          string
	ScreenName        string
	RelativeTweetTime string
	ReTweetCount      string
	FavoriteCount     string
	App               string
	TweetText         string
}

// OutputConfig is the configuration for outputting text from a tweet
type OutputConfig struct {
	MentionHighlightColor string
	HashtagHighlightColor string
	Highlight             bool
}

// TemplateOutput returns a TweetTemplateOutput based on the given tweet,
// this object should be used with the template library as an object for execution
func (t *Tweet) TemplateOutput(config OutputConfig) TweetTemplateOutput {
	return TweetTemplateOutput{
		CreatedAt:         t.CreatedAt,
		UserName:          t.User.Name,
		ScreenName:        t.User.ScreenName,
		RelativeTweetTime: t.RelativeTweetTime(),
		ReTweetCount:      strconv.Itoa(t.ReTweetCount),
		FavoriteCount:     strconv.Itoa(t.FavoriteCount),
		App:               util.ExtractAnchorText(t.Source),
		TweetText:         t.TweetText(config),
	}
}

// RelativeTweetTime returns a string output for display
//  if the tweet happened < 24 hours ago, then the relative time is 'XhYmZs ago'
//  otherwise the RelativeTweetTimeOutputLayout is used for time formatting.
func (t *Tweet) RelativeTweetTime() string {
	tstr := t.CreatedAt
	tm, err := time.Parse(CreatedAtTimeLayout, t.CreatedAt)
	if err == nil {
		since := time.Since(tm)
		if since < time.Hour*24 {
			tstr = since.Truncate(time.Second).String() + " ago"
		} else {
			tstr = tm.Format(RelativeTweetTimeOutputLayout)
		}
	}
	return tstr
}

func (t *Tweet) formatRetweetText(config OutputConfig) string {
	text := t.ReTweetedStatus.TweetText(config)
	screenName := fmt.Sprint("@", t.ReTweetedStatus.User.ScreenName)
	if config.Highlight {
		screenName = util.Colors.Colorize(config.MentionHighlightColor, screenName)
	}
	return fmt.Sprintf("RT %s: %s", screenName, text)
}

// TweetText returns the text of the tweet based on the given configuration
func (t *Tweet) TweetText(config OutputConfig) string {
	if t.ReTweetedStatus != nil {
		return t.formatRetweetText(config)
	}

	text := t.Text
	if len(t.FullText) > 0 {
		text = t.FullText
	}
	if config.Highlight {
		var hlents util.HighlightEntityList
		for _, ht := range t.Entities.HashTags {
			start, end := ht.Indices[0], ht.Indices[1]
			hlents = append(hlents, util.HighlightEntity{
				StartIdx: start,
				EndIdx:   end,
				Color:    config.HashtagHighlightColor})
		}

		for _, um := range t.Entities.UserMention {
			start, end := um.Indices[0], um.Indices[1]
			hlents = append(hlents, util.HighlightEntity{
				StartIdx: start,
				EndIdx:   end,
				Color:    config.MentionHighlightColor})
		}
		text = util.HighlightEntities(text, hlents)
	}
	return html.UnescapeString(text)
}

// SleepTime - from the twitter api
type SleepTime struct {
	Enabled   bool   `json:"enabled"`
	EndTime   *int64 `json:"end_time"`
	StartTime *int64 `json:"start_time"`
}

// TimeZone - from the twitter api
type TimeZone struct {
	Name       string `json:"name"`
	TzinfoName string `json:"tzinfo_name"`
	UtcOffset  int64  `json:"utc_offset"`
}

// PlaceType - from the twitter api
type PlaceType struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

// TrendLocation - from the twitter api
type TrendLocation struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
	ParentID    int64  `json:"parentid"`
	PlaceType   `json:"placeType"`
	URL         string `json:"url"`
	Woeid       int64  `json:"woeid"`
}

// AccountSettings - from the twitter api
type AccountSettings struct {
	AlwaysUseHTTPS           bool   `json:"always_use_https"`
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
