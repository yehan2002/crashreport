package ui

import "fmt"

type ui struct{ p *WebUI }

func (*ui) IsTerminal() bool                             { return false }
func (*ui) SetAutoComplete(complete func(string) string) {}
func (*ui) WantBrowser() bool                            { return false }
func (*ui) ReadLine(prompt string) (string, error)       { return "", nil }
func (u *ui) Print(v ...interface{})                     {}
func (u *ui) PrintErr(v ...interface{})                  { u.p.err = fmt.Errorf(fmt.Sprint(v...)) }
