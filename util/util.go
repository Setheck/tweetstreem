package util

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

var errUnsupportedPlatform = fmt.Errorf("unsupported platform")

var goos = runtime.GOOS
var startCommand = func(name string, args ...string) error { return exec.Command(name, args...).Start() }

// OpenBrowser opens the given url in a web browser.
// supports linux, windows, and darwin.
func OpenBrowser(url string) error {
	var name string
	var args []string
	switch goos {
	case "linux":
		name, args = "xdg-open", []string{url}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		name, args = "open", []string{url}
	default:
		return errUnsupportedPlatform
	}
	return startCommand(name, args...)
}

var notifier = signal.Notify // break out notifier for test

// Signal waits for os.Interrupt or syscall.SIGINT
func Signal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	notifier(ch, os.Interrupt, syscall.SIGINT)
	return ch
}

var anchorTextFind = regexp.MustCompile(`>(.+)<`)

// ExtractAnchorText extracts the node text from an html anchor element
func ExtractAnchorText(anchor string) string {
	found := anchorTextFind.FindStringSubmatch(anchor)
	if len(found) > 0 {
		return found[1]
	}
	return ""
}

// Stdin for testing TODO:(smt) need this?
var Stdin io.Reader = os.Stdin // replacable for testing

// SingleWordInput will scan Stdin and return the first word, discarding the rest of the input
// if for any reason, there is no first word, an empty string is returned
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

// FirstNumber return the first integer from the given list
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
	str = strings.TrimSpace(str)
	split := strings.Split(str, " ")
	if len(split) > 1 {
		return strings.ToLower(split[0]), split[1:]
	}
	return strings.ToLower(str), nil
}

// MustString panic on error
func MustString(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
