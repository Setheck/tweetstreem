package twitter

import (
	"fmt"
	"html"
	"sort"
	"strconv"
	"time"

	"github.com/Setheck/tweetstreem/util"
)

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
	Highlight             bool
}

func (t *Tweet) TemplateOutput(config OutputConfig) TweetTemplateOutput {
	return TweetTemplateOutput{
		UserName:          t.User.Name,
		ScreenName:        t.User.ScreenName,
		RelativeTweetTime: t.RelativeTweetTime(),
		ReTweetCount:      strconv.Itoa(t.ReTweetCount),
		FavoriteCount:     strconv.Itoa(t.FavoriteCount),
		App:               util.ExtractAnchorText(t.Source),
		TweetText:         t.TweetText(config),
	}
}

const (
	CreatedAtTimeLayout           = "Mon Jan 2 15:04:05 -0700 2006"
	RelativeTweetTimeOutputLayout = "01/02/2006 15:04:05"
)

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

func (t *Tweet) TweetText(config OutputConfig) string {
	text := ""
	if t.ReTweetedStatus != nil {
		text = t.ReTweetedStatus.TweetText(config)
		screenName := util.Colors.Colorize(config.MentionHighlightColor, "@"+t.ReTweetedStatus.User.ScreenName)
		return fmt.Sprintf("RT %s: %s", screenName, text)
	}
	if len(t.FullText) > 0 {
		text = t.FullText
	} else {
		text = t.Text
	}
	if config.Highlight {
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
		resultText += util.Colors.Colorize(entry.color, string(rtext[entry.startIdx:entry.endIdx]))
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
