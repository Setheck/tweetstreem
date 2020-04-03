package main

import (
	"fmt"
	"os"

	"github.com/Setheck/tweetstreem/app"
)

func main() {
	code := 0
	if err := app.NewTweetStreem().Run(); err != nil {
		fmt.Println("Error:", err)
		code = 1
	}
	os.Exit(code)
}
