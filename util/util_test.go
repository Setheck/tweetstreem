package util

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractAnchorText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no anchor", "testing", ""},
		{"simple anchor", "<a>test</a>", "test"},
		{"anchor with attribs", `<a class="test" id="123">test</a>`, "test"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ExtractAnchorText(test.input)
			if test.want != got {
				t.Fail()
			}
		})
	}
}

func TestFirstNumber(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
		ok   bool
	}{
		{"no number", []string{""}, 0, false},
		{"happy", []string{"1"}, 1, true},
		{"happy args", []string{"1", "2", "3"}, 1, true},
		{"second val", []string{"", "2", "3"}, 2, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := FirstNumber(test.args...)
			if ok != test.ok {
				t.Fail()
			}
			if test.want != got {
				t.Fail()
			}
		})
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		cmd   string
		args  []string
	}{
		{"happy", "help test", "help", []string{"test"}},
		{"single word", "oneword", "oneword", nil},
		{"single word uppercase", "ONEWORD", "oneword", nil},                       // BUG #18
		{"single all uppercase", "ONE TWO THREE", "one", []string{"TWO", "THREE"}}, // BUG #18
		{"empty", "", "", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd, args := SplitCommand(test.input)
			if cmd != test.cmd || !reflect.DeepEqual(args, test.args) {
				t.Fail()
			}
		})
	}
}

func TestSignal(t *testing.T) {
	sendCh := make(chan os.Signal, 1)

	// Replace the notifier to what we think it does very naively
	notifier = func(c chan<- os.Signal, sig ...os.Signal) {
		x := <-sendCh
		if x == os.Interrupt || x == syscall.SIGINT {
			c <- x
		} else {
			t.Fail() // Fail the test if we are expecting something that is not supported
		}
	}

	tests := []struct {
		name   string
		signal os.Signal
	}{
		{"interrupt", os.Interrupt},
		{"sigint", syscall.SIGINT},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sendCh <- test.signal // send test signal
			select {
			case out := <-Signal():
				// verify sent was received
				if out != test.signal {
					t.Fail()
				}
			case <-time.After(time.Millisecond * 10):
			}
		})
	}
}

func TestSingleWordInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"happy path", "123", "123"},
		{"multi word", "one two three", "one"},
		{"multi line", "one\n two\n three", "one"},
		{"no input", "", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Stdin is defined in util.go and defaults to os.Stdin
			Stdin = strings.NewReader(test.input)
			got := SingleWordInput()
			if test.want != got {
				t.Fail()
			}
		})
	}
}

func TestOpenBrowser(t *testing.T) {
	testURL := fmt.Sprintf("https://some.example.com?dt=%d", time.Now().Unix())
	tests := []struct {
		name string
		os   string
		url  string
	}{
		{"linux open", "linux", testURL},
		{"windows open", "windows", testURL},
		{"darwin open", "darwin", testURL},
		{"unsupported platform", "bsd", testURL},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			goos = test.os
			var expectedError error
			switch goos {
			case "linux":
				startCommand = func(name string, args ...string) error {
					assert.Equal(t, name, "xdg-open")
					assert.Equal(t, args, []string{test.url})
					return nil
				}
			case "windows":
				startCommand = func(name string, args ...string) error {
					assert.Equal(t, name, "rundll32")
					assert.Equal(t, args, []string{"url.dll,FileProtocolHandler", test.url})
					return nil
				}
			case "darwin":
				startCommand = func(name string, args ...string) error {
					assert.Equal(t, name, "open")
					assert.Equal(t, args, []string{test.url})
					return nil
				}
			default:
				startCommand = func(name string, args ...string) error {
					return nil
				}
				expectedError = errUnsupportedPlatform
			}
			err := OpenBrowser(test.url)
			if err != nil {
				assert.Equal(t, err, expectedError)
			}
		})
	}
}

func TestMustString(t *testing.T) {
	t.Run("panic if err", func(t *testing.T) {
		assert.Panics(t, func() {
			MustString("", assert.AnError)
		})
	})

	t.Run("returns string when no error", func(t *testing.T) {
		assert.Equal(t, "", MustString("", nil))
		assert.Equal(t, "testing", MustString("testing", nil))
	})
}
