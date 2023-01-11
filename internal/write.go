package internal

import "runtime"

// Header the header line to be used a crash report file.
// This text is written before the contents of the crash report
var Header = `crashreport
Use github.com/yehan2002/crashreport or open this file with any zip file viewer.
`

// Profiles the profiles to be included in the report
type Profiles uint8

// Profiles
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

func (p Profiles) Add(c *Config) {
	for i := 0; i < len(profiles); i++ {
		if p&0x1 == 1 {
			c.Profiles[profiles[i]] = struct{}{}
		}
		p = p >> 1
	}
}

// Config a struct containing config for creating a crash report.
type Config struct {
	Reason    []string
	Profiles  map[string]struct{}
	Files     []string
	NoStack   bool
	NoSysInfo bool
	MemStats  runtime.MemStats
}
