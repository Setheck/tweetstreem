package main

import (
	"os"

	"github.com/Setheck/tweetstreem/app"
)

func main() {
	code := app.NewTweetStreem().Run()
	os.Exit(code)
}
