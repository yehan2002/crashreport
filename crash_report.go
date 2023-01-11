package crashreport

import (
	"io"
	"os"

	"github.com/yehan2002/crashreport/internal"
)

// Profiles the profiles to be included in the report
type Profiles = internal.Profiles

// Profiles
const (
	ProfileHeap         Profiles = internal.ProfileHeap
	ProfileBlock                 = internal.ProfileBlock
	ProfileMutex                 = internal.ProfileMutex
	ProfileAllocs                = internal.ProfileAllocs
	ProfileGoroutines            = internal.ProfileGoroutines
	ProfileThreadCreate          = internal.ProfileThreadCreate

	ProfileAll = ProfileHeap | ProfileBlock | ProfileMutex | ProfileAllocs | ProfileGoroutines | ProfileThreadCreate
)

// CrashReport a crash report.
// The Write/WriteTo methods may be used multiple times.
type CrashReport struct{ c internal.Config }

// NewCrashReport creates a new crash report
func NewCrashReport(reason ...string) *CrashReport {
	c := &CrashReport{c: internal.Config{Reason: reason, Profiles: map[string]struct{}{}}}
	return c
}

// Include includes the given profiles in the crash report
func (c *CrashReport) Include(p Profiles) *CrashReport { p.Add(&c.c); return c }

// IncludeCustom includes a custom profile
func (c *CrashReport) IncludeCustom(name string) *CrashReport {
	c.c.Profiles[name] = struct{}{}
	return c
}

// IncludeFile includes the given file in the crash report.
// Any errors when including these files are silently ignored.
func (c *CrashReport) IncludeFile(path string) *CrashReport {
	c.c.Files = append(c.c.Files, path)
	return c
}

// NoStack excludes the stack from the crash report
func (c *CrashReport) NoStack() *CrashReport { c.c.NoStack = true; return c }

// NoSysInfo excludes system info from the crash report
func (c *CrashReport) NoSysInfo() *CrashReport { c.c.NoSysInfo = true; return c }

// Reason appends the given strings to the reason
func (c *CrashReport) Reason(s ...string) *CrashReport {
	c.c.Reason = append(c.c.Reason, s...)
	return c
}

func (c *CrashReport) Write(w io.Writer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}
			panic(r)
		}
	}()

	report, err := internal.Create(c.c)
	if err != nil {
		return err
	}

	err = report.Write(w)
	if err != nil {
		return err
	}

	return
}

// WriteTo writes the crash report to the given file
func (c *CrashReport) WriteTo(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	err = c.Write(f)
	if err != nil {
		return err
	}

	return f.Close()
}
