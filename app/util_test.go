package app

import (
	"reflect"
	"testing"
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
