package app

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"syscall"
)

func OpenBrowser(url string) error {
	fmt.Println("opening url in browser:", url)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func Signal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)
	return ch
}

var anchorTextFind = regexp.MustCompile(`>(.+)<`)

func ExtractAnchorText(anchor string) string {
	found := anchorTextFind.FindStringSubmatch(anchor)
	if len(found) > 0 {
		return found[1]
	}
	return ""
}

//func NumberString(args ...string) (int, string, bool) {
//	n := 0
//	str := ""
//	for _, a := range args {
//
//	}
//}

func FirstNumber(args ...string) (int, bool) {
	for _, a := range args {
		if n, err := strconv.Atoi(a); err == nil {
			return n, true
		}
	}
	return 0, false
}
