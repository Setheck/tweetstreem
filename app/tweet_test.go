package app

import (
	"testing"
)

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
	// TODO
}
