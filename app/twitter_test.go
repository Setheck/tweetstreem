package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestHighlight(t *testing.T) {
	t.SkipNow()
	data, err := ioutil.ReadFile("testData/tweets.json")
	if err != nil {
		t.Fatal(err)
	}
	var tweets []*Tweet
	if err := json.Unmarshal(data, &tweets); err != nil {
		t.Fatal(err)
	}

	for _, tw := range tweets {
		text := tw.TweetText(OutputConfig{
			MentionHighlightColor: "red",
			HashtagHighlightColor: "blue",
		}, true)
		fmt.Println(text)
	}
}
