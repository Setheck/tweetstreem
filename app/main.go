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

func failOnErr(args ...interface{}) {
	for _, arg := range args {
		if err, ok := arg.(error); ok && err != nil {
			log.Fatal(err)
		}
	}
}

// Run is the main entry point, returns result code
func Run() int {
	fmt.Print(appInfo())

	ts := NewTweetStreem(context.Background())
	if err := ts.LoadConfig(); err != nil {
		fmt.Println(err)
	}

	switch ParseFlags() {
	case version:
		return 0
	case client:
		failOnErr(ts.RemoteCall())
		return 0
	}

	// print pertinent config on start
	fmt.Printf("| auto-update | %s |\n",
		ts.TwitterConfiguration.PollTimeDuration())

	failOnErr(ts.StartSubsystems())

	ts.WaitForDone()

	// Shutdown Sequence
	failOnErr(ts.SaveConfig())

	fmt.Println("\n'till next time o/ ")
	return 0
}
