package crashreport

import (
	"io"
	"os"
	"runtime"

	"github.com/yehan2002/crashreport/internal"
)

//Profiles the profiles to be included in the report
type Profiles uint8

//Profiles
const (
	ProfileHeap Profiles = 1 << iota
	ProfileBlock
	ProfileMutex
	ProfileAllocs
	ProfileGoroutines
	ProfileThreadCreate

	ProfileAll = ProfileHeap | ProfileBlock | ProfileMutex | ProfileAllocs | ProfileGoroutines | ProfileThreadCreate
)

var profiles = [...]string{"heap", "block", "mutex", "allocs", "goroutine", "threadcreate"}

//CrashReport a crash report.
//The Write/WriteTo methods may be used multiple times.
type CrashReport struct {
	reason    []string
	profiles  map[string]int
	files     []string
	noStack   bool
	noSysInfo bool
	mem       runtime.MemStats
}

//NewCrashReport creates a new crash report
func NewCrashReport(reason ...string) *CrashReport {
	c := &CrashReport{reason: reason, profiles: map[string]int{}}
	runtime.ReadMemStats(&c.mem)
	return c
}

//Include includes the given profiles in the crash report
func (c *CrashReport) Include(p Profiles) *CrashReport {
	for i := 0; i < len(profiles); i++ {
		if p&0x1 == 1 {
			c.profiles[profiles[i]] = 1
		}
		p = p >> 1
	}
	return c
}

//IncludeCustom includes a custom profile
func (c *CrashReport) IncludeCustom(name string) *CrashReport { c.profiles[name] = 1; return c }

//IncludeFile includes the given file in the crash report.
//Any errors when including these files are silently ignored.
func (c *CrashReport) IncludeFile(file string) *CrashReport {
	c.files = append(c.files, file)
	return c
}

//NoStack excludes the stack from the crash report
func (c *CrashReport) NoStack() *CrashReport { c.noStack = true; return c }

//NoSysInfo excludes system info from the crash report
func (c *CrashReport) NoSysInfo() *CrashReport { c.noSysInfo = true; return c }

//Reason appends the given strings to the reason
func (c *CrashReport) Reason(s ...string) *CrashReport { c.reason = append(c.reason, s...); return c }

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

	cw := internal.CreateWriter(w)
	cw.Reason(c.reason)
	for name := range c.profiles {
		cw.Profile(name)
	}
	for _, file := range c.files {
		cw.File(file)
	}
	if !c.noStack {
		cw.Stack()
	}
	if !c.noSysInfo {
		cw.SysInfo()
	}
	cw.Close()
	return
}

//WriteTo writes the crash report to the given file
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
