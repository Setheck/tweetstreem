package util

import (
	"github.com/atotto/clipboard"
)

// Clipper is the clipboard helper interface
type Clipper interface {
	ReadAll() (string, error)
	WriteAll(s string) error
}

// ClipboardHelper is the default helper which interacts with the clipboard
var ClipboardHelper Clipper = defaultClipper{}

type defaultClipper struct{}

// ReadAll read string from clipboard
func (c defaultClipper) ReadAll() (string, error) {
	return clipboard.ReadAll()
}

// WriteAll overwrites the clipboard with the given string
func (c defaultClipper) WriteAll(s string) error {
	return clipboard.WriteAll(s)
}
