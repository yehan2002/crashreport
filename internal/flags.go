package internal

import (
	"github.com/google/pprof/driver"
)

type fakeFlags struct {
	file string
}

func (*fakeFlags) Bool(o string, d bool, c string) *bool {
	var s bool = d
	switch o {
	case "no_browser":
		s = true
	}
	return &s
}

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

//NewFakeFlag flag
func NewFakeFlag(file string) driver.FlagSet {
	return &fakeFlags{file}
}
