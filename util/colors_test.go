package util

import (
	"fmt"
	"runtime"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestHighlightEntities(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		entities HighlightEntityList
		want     string
	}{
		{
			"invalid  entity startidx",
			"test one two three",
			[]HighlightEntity{{-1, 0, "blue"}},
			"test one two three",
		},
		{
			"invalid  entity endidx",
			"test one two three",
			[]HighlightEntity{{2, 1, "blue"}},
			"test one two three",
		},
		{
			"no entity highlight",
			"test one two three",
			[]HighlightEntity{},
			"test one two three",
		},
		{
			"single entity highlight",
			"test one two three",
			[]HighlightEntity{{5, 8, "red"}},
			fmt.Sprint("test ", Colors.Colorize("red", "one"), " two three"),
		},
		{
			"multi entity highlight",
			"test one two three",
			[]HighlightEntity{
				{5, 8, "red"},
				{9, 12, "blue"},
			},
			fmt.Sprint("test ",
				Colors.Colorize("red", "one"),
				" ",
				Colors.Colorize("blue", "two"),
				" three"),
		},
		{
			"whole text highlight",
			"test one two three",
			[]HighlightEntity{{0, 18, "red"}},
			Colors.Colorize("red", "test one two three"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := HighlightEntities(test.text, test.entities)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestHighlightEntityList_Sortable(t *testing.T) {
	first := HighlightEntity{StartIdx: 0}
	second := HighlightEntity{StartIdx: 5}
	third := HighlightEntity{StartIdx: 10}
	sortedExpectation := HighlightEntityList{first, second, third}

	inOrder := HighlightEntityList{first, second, third}
	sort.Sort(inOrder)
	assert.Equal(t, sortedExpectation, inOrder)

	outOfOrder := HighlightEntityList{second, third, first}
	sort.Sort(outOfOrder)
	assert.Equal(t, sortedExpectation, outOfOrder)

	reversed := HighlightEntityList{third, second, first}
	sort.Sort(reversed)
	assert.Equal(t, sortedExpectation, reversed)
}
