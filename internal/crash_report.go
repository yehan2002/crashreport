package internal

import (
	"runtime"
	"runtime/debug"
	"time"
)

// startTime the time the program started running at
var startTime = time.Now()

type CrashReport struct {
	Profiles []*Profile

	SysInfo  *SysInfo
	Memstats *runtime.MemStats
	Build    *debug.BuildInfo

	Reason string
	Stack  string

	Files []string
}

// SysInfo system info
type SysInfo struct {
	Arch      string
	OS        string
	Compiler  string
	GoVersion string

	CPU        int
	Goroutines int
	Threads    int
	MaxCPU     int

	Mem         uint64
	GcPauseNs   uint64
	LastGcPause uint64
	GcCPU       float64

	Time        time.Time
	TimeStart   time.Time
	TimeRunning time.Duration
}

func getSysInfo() *SysInfo {
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
