package app

import (
	"context"
	"fmt"
	"log"
)

var (
	Version = "0.0.1"
	Commit  = "dev"
	Built   = "0"
)

func appInfo() string {
	return "~~~~~~~~~~~~~~~\n" +
		"~~ Tweet %%%%%%" + fmt.Sprintln(" version: ", Version) +
		"~~~~~ Streem %%" + fmt.Sprintln(" commit:  ", Commit) +
		"~~~~~~~~~~~~~~~" + fmt.Sprintln(" built:   ", Built)
}

// Run is the main entry point, returns result code
func Run() int {
	fmt.Println(appInfo())

	ts := NewTweetStreem(context.Background())
	if err := ts.LoadConfig(); err != nil {
		fmt.Println(err)
	}

	if err := ts.ParseFlags(); err != nil {
		log.Fatal(err)
	}

	// print pertinent config on start
	fmt.Printf("| auto-update | %s |\n",
		ts.TwitterConfiguration.PollTimeDuration())

	if err := ts.StartSubsystems(); err != nil {
		log.Fatal(err)
	}

	<-ts.ctx.Done()

	// Shutdown Sequence
	if err := ts.SaveConfig(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("\n'till next time o/ ")
	return 0
}
