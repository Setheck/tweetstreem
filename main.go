package main

import (
	"os"

	"github.com/Setheck/tweetstreem/app"
)

func main() {
	code := app.Run()
	os.Exit(code)
}
