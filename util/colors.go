package util

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type terminalColors map[string]string

// Colors is the exported set of terminal colors
var Colors = terminalColors{
	"reset":   "\033[0m",
	"black":   "\033[30m",
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"gray":    "\033[37m",
	"white":   "\033[97m",
}

func init() {
	// sry winderz
	if runtime.GOOS == "windows" {
		Colors = terminalColors{}
	}
}

// Code returns the terminal color code for the given color name
func (c terminalColors) Code(color string) string {
	lcolor := strings.ToLower(color)
	if val, ok := c[lcolor]; ok {
		return val
	}
	return ""
}

// Colorize returns the given string wrapped in the appropriately named color code and the 'reset' color code.
func (c terminalColors) Colorize(color, str string) string {
	return fmt.Sprint(c.Code(color), str, c.Code("reset"))
}

// HighlightEntity is a structure that represents sections in a string to be highlighted
type HighlightEntity struct {
	StartIdx int
	EndIdx   int
	Color    string
}

// HighlightEntityList a sortable list of HighlightEntity
type HighlightEntityList []HighlightEntity

func (l HighlightEntityList) Len() int {
	return len(l)
}
func (l HighlightEntityList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l HighlightEntityList) Less(i, j int) bool {
	return l[i].StartIdx < l[j].StartIdx
}

// HighlightEntities attempts to colorize the given text, with the given HighlightEntityList
func HighlightEntities(text string, hlist HighlightEntityList) string {
	sort.Sort(hlist)
	rtext := []rune(text)
	resultText := ""
	curIdx := 0
	for _, entry := range hlist {
		if entry.StartIdx >= curIdx && entry.EndIdx <= len(text) && entry.StartIdx <= entry.EndIdx {
			resultText += string(rtext[curIdx:entry.StartIdx])
			resultText += Colors.Colorize(entry.Color, string(rtext[entry.StartIdx:entry.EndIdx]))
			curIdx = entry.EndIdx
		}
	}
	resultText += string(rtext[curIdx:])
	return resultText
}
