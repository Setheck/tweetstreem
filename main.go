package main

import (
	"log"

	"github.com/Setheck/tweetstream/app"
)

func main() {
	tw := app.NewTweetStream()
	if err := tw.Init(); err != nil {
		log.Fatal(err)
	}
	tw.WatchTerminal()
	if err := tw.Stop(); err != nil {
		log.Fatal(err)
	}
}
