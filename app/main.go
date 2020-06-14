package app

import (
	"context"
	"flag"
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

type RunMode string

const (
	version RunMode = "version"
	client  RunMode = "client"
	normal  RunMode = "normal"
)

func ParseFlags() RunMode {
	verFlg := flag.Bool("v", false, "version")
	clientFlg := flag.Bool("c", false, "client input")
	flag.Parse()

	switch {
	case *verFlg:
		return version
	case *clientFlg:
		return client
	}
	return normal
}

// Run is the main entry point, returns result code
func Run() int {
	fmt.Println(appInfo())

	ts := NewTweetStreem(context.Background())
	if err := ts.LoadConfig(); err != nil {
		fmt.Println(err)
	}

	switch ParseFlags() {
	case version:
		return 0
	case client:
		if err := ts.RemoteCall(); err != nil {
			log.Fatal(err)
		}
		return 0
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
