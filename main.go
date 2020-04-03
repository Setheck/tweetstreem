package main

import (
	"fmt"
	"os"

	"github.com/Setheck/tweetstream/app"
)

func main() {
	code := 0
	if err := app.NewTweetStream().Run(); err != nil {
		fmt.Println("Error:", err)
		code = 1
	}
	os.Exit(code)
}
