package ui

import (
	"fmt"
	"os"
)

type profUI struct{}

func (*profUI) IsTerminal() bool                             { return false }
func (*profUI) SetAutoComplete(complete func(string) string) {}
func (*profUI) WantBrowser() bool                            { return false }
func (*profUI) ReadLine(prompt string) (string, error)       { return "", nil }
func (u *profUI) Print(v ...interface{})                     {}
func (u *profUI) PrintErr(v ...interface{})                  { fmt.Fprint(os.Stderr, v...) }
