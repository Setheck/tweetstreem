package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
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

func ParseFlags(port int) {
	b := flag.Bool("c", false, "client input")
	flag.Parse()

	if *b {
		client := NewRemoteClient(fmt.Sprintf(":%d", port))
		input := strings.Join(flag.Args(), " ")
		if err := client.RpcCall(input); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
}

// Run is the main entry point, returns result code
func Run() int {
	ts := NewTweetStreem()
	loadConfig(ts)
	ParseFlags(ts.ApiPort)

	fmt.Println(Banner)
	fmt.Println("polling every:", ts.PollTime.Truncate(time.Second).String())

	if err := ts.initTwitter(); err != nil {
		fmt.Println("Error:", err)
		return 1
	}
	fmt.Println("Api")
	if ts.EnableApi {
		ts.initApi()
	}

	go ts.watchTerminalInput()
	go ts.echoOnPoll()
	go ts.consumeInput()
	go ts.signalWatcher()

	if ts.AutoHome {
		_ = ts.homeTimeline()
	}
	<-ts.ctx.Done()
	saveConfig(ts)
	fmt.Println("\n'till next time o/ ")
	return 0
}
