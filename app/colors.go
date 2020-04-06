package app

import (
	"fmt"
	"runtime"
	"strings"
)

type terminalColors map[string]string

var Colors = terminalColors{
	"reset":  "\033[0m",
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"blue":   "\033[34m",
	"purple": "\033[35m",
	"cyan":   "\033[36m",
	"gray":   "\033[37m",
	"white":  "\033[97m",
}

func init() {
	// sry winderz
	if runtime.GOOS == "windows" {
		Colors = terminalColors{}
	}
}

func (c terminalColors) Code(color string) string {
	lcolor := strings.ToLower(color)
	if val, ok := c[lcolor]; ok {
		return val
	}
	return ""
}

func (c terminalColors) Colorize(color, str string) string {
	return fmt.Sprint(c.Code(color), str, c.Code("reset"))
}
