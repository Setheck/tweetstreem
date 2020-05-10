package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func OpenBrowser(url string) error {
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

var Notifier = signal.Notify // break out notifier for test

func Signal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	Notifier(ch, os.Interrupt, syscall.SIGINT)
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

var Stdin io.Reader = os.Stdin // replacable for testing

func SingleWordInput() string {
	stdin := bufio.NewScanner(Stdin)
	if stdin.Scan() {
		fields := strings.Fields(stdin.Text())
		if len(fields) > 0 {
			return fields[0]
		}
	}
	return ""
}

func FirstNumber(args ...string) (int, bool) {
	for _, a := range args {
		if n, err := strconv.Atoi(a); err == nil {
			return n, true
		}
	}
	return 0, false
}

// SplitCommand takes a string and returns command and arguments
func SplitCommand(str string) (string, []string) {
	str = strings.ToLower(str)
	str = strings.TrimSpace(str)
	split := strings.Split(str, " ")
	if len(split) > 1 {
		return split[0], split[1:]
	}
	if len(split) > 0 {
		return split[0], nil
	}
	return "", nil
}
