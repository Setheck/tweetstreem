package main

import (
	"fmt"
	"log"

	"github.com/Setheck/tweetstream/app"
)

func main() {
	tw := app.NewTweetStream()
	if err := tw.Init(); err != nil {
		log.Fatal(err)
	}

	for {
		var input string
		fmt.Scanln(&input)
		fmt.Println("read:", input)

		switch {
		case input == "exit":
			if err := tw.Stop(); err != nil {
				log.Println(err)
			}
			return
		case input == "home":
			tw.GetHomeTimeline()
		}
		fmt.Printf("\nPrompt:")
	}

}
