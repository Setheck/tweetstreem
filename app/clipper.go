package app

import (
	"github.com/atotto/clipboard"
)

type Clipper interface {
	ReadAll() (string, error)
	WriteAll(s string) error
}

var ClipboardHelper Clipper = DefaultClipper{}

type DefaultClipper struct{}

// ReadFromClipboard is a helper to overwrite the args with the clipboard
// falls back to given args,
func (c DefaultClipper) ReadAll() (string, error) {
	return clipboard.ReadAll()
}

func (c DefaultClipper) WriteAll(s string) error {
	return clipboard.WriteAll(s)
}
