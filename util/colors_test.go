package util

import (
	"fmt"
	"runtime"
	"testing"
)

func TestTerminalColors_Code(t *testing.T) {
	tests := []struct {
		color string
		code  string
	}{
		{"reset", "\033[0m"},
		{"black", "\033[30m"},
		{"red", "\033[31m"},
		{"green", "\033[32m"},
		{"yellow", "\033[33m"},
		{"blue", "\033[34m"},
		{"magenta", "\033[35m"},
		{"cyan", "\033[36m"},
		{"gray", "\033[37m"},
		{"white", "\033[97m"},
	}
	for _, test := range tests {
		t.Run(test.color, func(t *testing.T) {
			want := test.code
			if runtime.GOOS == "windows" {
				want = ""
			}
			if Colors.Code(test.color) != want {
				t.Fail()
			}
		})
	}
}

func TestTerminalColors_Colorize(t *testing.T) {
	tests := []struct {
		color string
	}{
		{"reset"},
		{"black"},
		{"red"},
		{"green"},
		{"yellow"},
		{"blue"},
		{"magenta"},
		{"cyan"},
		{"gray"},
		{"white"},
	}
	testString := "this is a test string"
	for _, test := range tests {
		t.Run(test.color, func(t *testing.T) {
			got := Colors.Colorize(test.color, testString)
			if runtime.GOOS == "windows" {
				if got != testString {
					t.Fail()
				}
			} else {
				code := Colors.Code(test.color)
				reset := Colors.Code("reset")
				want := fmt.Sprint(code, testString, reset)
				fmt.Println(want) // For visual color inspection
				if got != want {
					t.Fail()
				}
			}
		})
	}
}
