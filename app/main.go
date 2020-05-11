package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	ts := NewTweetStreem(nil)
	loadConfig(ts)
	ParseFlags(ts)

	fmt.Println(Banner)
	fmt.Println("polling every:", ts.TwitterConfiguration.PollTimeDuration())

	if err := ts.initTwitter(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	if ts.EnableApi {
		fmt.Println("api server enabled on port:", ts.ApiPort)
		ts.initApi()
	}

	go ts.watchTerminalInput()
	go ts.echoOnPoll()
	go ts.consumeInput()
	go ts.outputPrinter()
	go ts.signalWatcher()

	if ts.AutoHome {
		_ = ts.homeTimeline()
	}
	<-ts.ctx.Done()
	saveConfig(ts)
	fmt.Println("\n'till next time o/ ")
	return 0
}
