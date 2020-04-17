TweetStreem
---
A cross platform twitter client for the terminal. 
Inspired heavily by [rainbowstream](https://github.com/orakaro/rainbowstream)

### Actions
* config - show the current configuration
* p,pause - pause the stream
* r,resume - resume the stream
* v,version - print tweetstreem version
* o,open - open the link in the selected tweet (optionally provide 0 based index)
* b,browse - open the selected tweet in a browser
* rt,retweet - retweet the selected tweet
* urt,unretweet - uretweet the selected tweet
* li,like - like the selected tweet
* ul,unlike - unlike the selected tweet
* reply <id> <status> - reply to the tweet id (requires user mention, and confirmation)
* cbreply <id> - reply to tweet id with clipboard contents (requires confirmation)
* t,tweet <status> - create a new tweet and post (requires confirmation)
* me - view your recent tweets
* home - view your default timeline
* q,quit,exit - exit tweetstreem.
* h,help - this help menu :D

### Configuration
Tweetstream will create `$HOME/.tweetstreem.json`
Example of default configuration
```
{
  "config": {
    "twitterConfiguration": {
      "pollTime": 120000000000,
      "userToken": "*****",
      "userSecret": "*****"
    },
    "tweetTemplate": "\n{{ .UserName | color \"cyan\" }} {{ \"@\" | color \"green\" }}{{ .ScreenName | color \"green\" }} {{ .RelativeTweetTime | color \"magenta\" }}\nid:{{ .Id }} {{ \"rt:\" | color \"cyan\" }}{{ .ReTweetCount | color \"cyan\" }} {{ \"♥:\" | color \"red\" }}{{ .FavoriteCount | color \"red\" }} via {{ .App | color \"blue\" }}\n{{ .TweetText }}\n",
    "templateOutputConfig": {
      "MentionHighlightColor": "blue",
      "HashtagHighlightColor": "magenta"
    },
    "enableApi": false,
    "apiPort": 8080,
    "autoHome": false
  }
}
```

### Templating
Output of tweets is based on go templates and some home grown helpers.
The default Template is:

```

{{ .UserName | color "cyan" }} {{ "@" | color "green" }}{{ .ScreenName | color "green" }} {{ .RelativeTweetTime | color "magenta" }}
id:{{ .Id }} {{ "rt:" | color "cyan" }}{{ .ReTweetCount | color "cyan" }} {{ "♥:" | color "red" }}{{ .FavoriteCount | color "red" }} via {{ .App | color "blue" }}
{{ .TweetText }}

  ```

Template Helpers that exist are
`color <colorstr> <text to colorize>`

*Note Windows terminal does not support colors*

Available colors are 
* black
* red
* green
* yellow
* blue
* magenta
* cyan
* gray
* white

