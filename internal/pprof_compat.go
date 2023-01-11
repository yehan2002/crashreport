package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/google/pprof/profile"
)

// This file contains struct used to make pprof work with the web ui.

type profUI struct{}

func (*profUI) IsTerminal() bool                             { return false }
func (*profUI) SetAutoComplete(complete func(string) string) {}
func (*profUI) WantBrowser() bool                            { return false }
func (*profUI) ReadLine(prompt string) (string, error)       { return "", nil }
func (u *profUI) Print(v ...interface{})                     {}
func (u *profUI) PrintErr(v ...interface{})                  { fmt.Fprint(os.Stderr, v...) }

type fakeFlags struct {
	file string
}

func (*fakeFlags) Bool(o string, d bool, c string) *bool { return new(bool) }

func (*fakeFlags) String(o, d, c string) *string {
	var s string = d
	switch o {
	case "http":
		s = ":0000"
	}
	return &s
}

func (*fakeFlags) Int(o string, d int, c string) *int                   { return &d }
func (*fakeFlags) Float64(o string, d float64, c string) *float64       { return &d }
func (*fakeFlags) BoolVar(b *bool, o string, d bool, c string)          { *b = d }
func (*fakeFlags) IntVar(i *int, o string, d int, c string)             { *i = d }
func (*fakeFlags) Float64Var(f *float64, o string, d float64, c string) { *f = d }
func (*fakeFlags) StringVar(s *string, o, d, c string)                  { *s = d }
func (*fakeFlags) StringList(o, d, c string) *[]*string                 { return &[]*string{&d} }
func (*fakeFlags) ExtraUsage() string                                   { return "" }
func (*fakeFlags) AddExtraUsage(eu string)                              {}
func (f *fakeFlags) Parse(usage func()) []string                        { return []string{f.file} }

type fetcher struct{ P *profile.Profile }

func (f *fetcher) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	return f.P, "", nil
}
