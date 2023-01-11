package internal

import (
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
)

// startTime the time the program started running at.
var startTime = time.Now()

// CrashReport a crash report
type CrashReport struct {
	// Profiles profiles included in the crash report
	Profiles []*Profile

	// SysInfo contains information about the system the process was running in.
	// This will be nil if [Config.NoSysInfo] is true or if the system.json does
	// not exist in the crash report file.
	SysInfo *SysInfo
	// Memstats memory usage statistics of the program.
	// This will be nil if memstats.json does not exist in the crash report file.
	Memstats *runtime.MemStats
	// Build contains build info embedded in the binary of the program.
	// This will be nil if the build.json does not exist in the crash report file.
	Build *debug.BuildInfo

	// Reason the reason the program crashed.
	Reason string
	// Stack the full stack trace of the program
	Stack string

	// Files extra files included in the crash report.
	Files []string
}

// SysInfo contains information about the system the process was running in.
type SysInfo struct {
	// Arch the running programs architecture target.
	// This is the value of [runtime.GOARCH]
	Arch string
	// OS the running programs operating system target.
	// This is the value of [runtime.OS]
	OS string
	// Compiler is the name of the compiler toolchain that built the running binary.
	Compiler string
	// GoVersion is the Go tree's version string.
	// It is either the commit hash and date at the time of the build or,
	// when possible, a release tag like "go1.3".
	GoVersion string
	// CPU is the number of logical CPUs usable by the process.
	//
	// The set of available CPUs is checked by querying the operating system
	// at process startup. Changes to operating system CPU allocation after
	// process startup are not reflected.
	CPU int
	// Goroutine the number of goroutines that existed at the time the crash report was created.
	Goroutines int
	// Thread the number of os the process was using when the crash report was created.
	Threads int
	// MaxCPU is the value of GOMAXPROCS.
	MaxCPU int

	// Time is the time when the crash report was created.
	Time time.Time
	// TimeStart is the time the program started running at.
	TimeStart time.Time
	// TimeRunning is the total amount of time the program was running.
	TimeRunning time.Duration
}

func newSysInfo() *SysInfo {
	threads, _ := runtime.ThreadCreateProfile([]runtime.StackRecord{})
	return &SysInfo{
		Arch:       runtime.GOARCH,
		OS:         runtime.GOOS,
		Compiler:   runtime.Compiler,
		GoVersion:  runtime.Version(),
		CPU:        runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		Threads:    threads,
		MaxCPU:     runtime.GOMAXPROCS(-1),

		Time:        time.Now(),
		TimeStart:   startTime,
		TimeRunning: time.Since(startTime),
	}
}

// Profile a profile
type Profile struct {
	profile []byte
	name    string
}

func (p *Profile) URL() string  { return strings.ToLower(p.name) }
func (p *Profile) Name() string { return p.name }

func (p *Profile) Profile() (*profile.Profile, error) {
	return profile.ParseData(p.profile)
}

func (p *Profile) ProfileBytes() []byte { return p.profile }

func (p *Profile) Register(mux *http.ServeMux) error {
	prof, err := p.Profile()
	if err != nil {
		return err
	}

	return driver.PProf(&driver.Options{HTTPServer: func(d *driver.HTTPServerArgs) error {
		for path, handler := range d.Handlers {
			u, err := url.JoinPath("/profile", p.URL(), path)
			if err != nil {
				return err
			}

			mux.Handle(u, handler)
		}
		return nil
	}, UI: &profUI{}, Flagset: &fakeFlags{}, Fetch: &fetcher{P: prof}})
}

func NewProfile(name string, prof []byte) *Profile {
	return &Profile{profile: prof, name: strings.Title(name)}
}
