package twitter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//func TestHighlight(t *testing.T) {
//	t.SkipNow()
//	data, err := ioutil.ReadFile("testData/tweets.json")
//	if err != nil {
//		t.Fatal(err)
//	}
//	var tweets []*Tweet
//	if err := json.Unmarshal(data, &tweets); err != nil {
//		t.Fatal(err)
//	}
//
//	for _, tw := range tweets {
//		text := tw.TweetText(OutputConfig{
//			MentionHighlightColor: "red",
//			HashtagHighlightColor: "blue",
//			Highlight:             true,
//		})
//		fmt.Println(text)
//	}
//}

func TestTweet_HtmlLink(t *testing.T) {
	tweet := &Tweet{
		IDStr: "12345",
		User:  User{ScreenName: "test_user"},
	}

	want := "https://twitter.com/test_user/status/12345"
	got := tweet.HtmlLink()
	if want != got {
		t.Fail()
	}
}

func TestTweet_Links(t *testing.T) {
	tests := []struct {
		name  string
		urls  []Url
		media []Media
	}{
		{
			"retrieve urls",
			[]Url{{ExpandedUrl: "https://url1.com"}},
			[]Media{{MediaUrl: "https://url2.com"}},
		},
		{
			"nothing to retrieve",
			[]Url{},
			[]Media{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &Tweet{
				Entities: Entities{
					Urls:  test.urls,
					Media: test.media,
				},
			}

			links := tweet.Links()
			if len(links) < 1 {
				assert.Empty(t, test.urls)
				assert.Empty(t, test.media)
			}
			for _, u := range test.urls {
				assert.Contains(t, links, u.ExpandedUrl)
			}
			for _, u := range test.media {
				assert.Contains(t, links, u.MediaUrl)
			}
		})
	}
}

func TestTweet_TemplateOutput(t *testing.T) {
	tests := []struct {
		name       string
		outputConf OutputConfig
	}{
		{"default colors", OutputConfig{}},
		{"blue mentions", OutputConfig{MentionHighlightColor: "blue"}},
		{"blue hashtags", OutputConfig{HashtagHighlightColor: "blue"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userName, screenName := "user name", "screen name"
			reTweetCount := "0"
			favoriteCount := "0"
			app := "this is a test"

			tweet := &Tweet{
				User: User{
					Name:       userName,
					ScreenName: screenName,
				},
				Source: `<a href="http://example.com">this is a test</a>`,
			}
			relativeTweetTime := tweet.RelativeTweetTime()
			tweetText := tweet.TweetText(test.outputConf)

			output := tweet.TemplateOutput(test.outputConf)

			assert.Equal(t, userName, output.UserName)
			assert.Equal(t, screenName, output.ScreenName)
			assert.Equal(t, relativeTweetTime, output.RelativeTweetTime)
			assert.Equal(t, reTweetCount, output.ReTweetCount)
			assert.Equal(t, favoriteCount, output.FavoriteCount)
			assert.Equal(t, app, output.App)
			assert.Equal(t, tweetText, output.TweetText)
		})
	}
}

func TestTweet_RelativeTweetTime_BadCreatedAt(t *testing.T) {
	want := ""
	tweet := &Tweet{CreatedAt: ""}
	assert.Equal(t, want, tweet.RelativeTweetTime())

	want = "not a date 1234332"
	tweet = &Tweet{CreatedAt: "not a date 1234332"}
	assert.Equal(t, want, tweet.RelativeTweetTime())
}

func TestTweet_RelativeTweetTime(t *testing.T) {
	now := time.Now()
	fiveMinPast := now.Add(-5 * time.Minute)
	tenMinThirtySecPast := now.Add(-10 * time.Minute).Add(-30 * time.Second)
	twentyThreeHoursPast := now.Add(-23 * time.Hour)
	twentyFourHoursPast := now.Add(-24 * time.Hour)
	tests := []struct {
		name      string
		createdAt time.Time
		want      string
	}{
		{"five min ago", fiveMinPast, "5m0s ago"},
		{"10m30s min ago", tenMinThirtySecPast, "10m30s ago"},
		{"24 hours ago", twentyThreeHoursPast, "23h0m0s ago"},
		{"24 hours ago", twentyFourHoursPast, twentyFourHoursPast.Format(RelativeTweetTimeOutputLayout)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &Tweet{CreatedAt: test.createdAt.Format(CreatedAtTimeLayout)}
			got := tweet.RelativeTweetTime()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestTweet_TweetText(t *testing.T) {
	highLight := OutputConfig{MentionHighlightColor: "blue", HashtagHighlightColor: "blue"}
	noHighLight := OutputConfig{}

	tests := []struct {
		name       string
		text       string
		outputConf OutputConfig
		highlight  bool
	}{
		{"no colors", "", noHighLight, false},
		{"no colors", "", highLight, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tweet := &Tweet{}
			tweet.Text = test.text
			twText := tweet.TweetText(test.outputConf)
			assert.Equal(t, "", twText)

			tweet.FullText = test.text
			twText = tweet.TweetText(test.outputConf)
			assert.Equal(t, "", twText)
		})
	}
}
