package app

import (
	"fmt"
	"time"
)

type Twitter struct {
}

func (t *Twitter) HomeTimeline() {

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
		tstr = time.Since(tm).Truncate(time.Minute).String()
	}
	return fmt.Sprintf("%s @%s %s", t.User.Name, t.User.ScreenName, tstr)
}

func (t *Tweet) StatusString() string {
	return fmt.Sprintf("rt:%d ♥:%d", t.ReTweetCount, t.FavoriteCount)
}

func (t *Tweet) String() string {
	return t.Text
}
