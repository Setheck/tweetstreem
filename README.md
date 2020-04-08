TweetStreem
---

A cross platform twitter client for the terminal.

###Templating
Output of tweets is based on go templates and some home grown helpers.
The default Template is

```
{{ .UserName | color "cyan" }} {{ .ScreenName | color "green" }} {{ .RelativeTweetTime | color "magenta" }}
  id:{{ .Id }} {{ "rt:" | color "cyan" }}{{ .ReTweetCount | color "cyan" }} {{ "♥:" | color "red" }}{{ .FavoriteCount | color "red" }} via {{ .App | color "blue" }}
  {{ .TweetText }}
  
  ```

Template Helpers that exist are
`color <colorstr> <text to colorize>`

*Note Windows terminal does not support colors*