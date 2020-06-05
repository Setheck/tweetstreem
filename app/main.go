package app

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Setheck/tweetstreem/twitter"
)

var (
	Version = "0.0.1"
	Commit  = "dev"
	Built   = "0"
)

const Banner = `
~~~~~~~~~~~~~~~~
~~Tweet
~~   Streem
~~~~~~~~~~~~~~~~`

func version() string {
	return fmt.Sprintln(Banner) +
		fmt.Sprintln("version:", Version) +
		fmt.Sprintln(" commit:", Commit) +
		fmt.Sprintln("  built:", Built)
}

var ExitFn = os.Exit // replaceable for test
func ParseFlags(ts *TweetStreem) {
	clientMode := flag.Bool("c", false, "client input")
	flag.Parse()

	if *clientMode {
		client := NewRemoteClient(ts, fmt.Sprintf("%s:%d", ts.ApiHost, ts.ApiPort))
		input := strings.Join(flag.Args(), " ")
		if err := client.RpcCall(input); err != nil {
			log.Fatal(err)
		}
		ExitFn(0)
	}
}

// Run is the main entry point, returns result code
func Run() int {
	ts := NewTweetStreem(context.Background())
	if err := LoadConfig(ts); err != nil {
		fmt.Println(err)
	}
	ParseFlags(ts)

	fmt.Println(Banner)
	fmt.Println("polling every:", ts.TwitterConfiguration.PollTimeDuration())
	if err := ts.ParseTemplate(); err != nil {
		fmt.Println(err)
	}

	ts.twitter = twitter.NewDefaultClient(*ts.TwitterConfiguration)
	if err := ts.twitter.Authorize(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}

	if ts.EnableApi {
		fmt.Println("api server enabled on port:", ts.ApiPort)
		ts.initApi()
	}

	go ts.watchTerminalInput()
	go ts.pollAndEcho()
	go ts.consumeInput()
	go ts.outputPrinter()
	go ts.waitForDone()

	if ts.AutoHome {
		_ = ts.homeTimeline()
	}
	<-ts.ctx.Done()
	conf := ts.twitter.Configuration()
	ts.TwitterConfiguration = &conf
	if err := SaveConfig(ts); err != nil {
		fmt.Println(err)
	}
	fmt.Println("\n'till next time o/ ")
	return 0
}
