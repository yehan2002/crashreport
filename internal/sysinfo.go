package internal

import (
	"runtime"
	"runtime/debug"
	"time"
)

var startTime = time.Now()

//SysInfo system info
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

	buildInfo *debug.BuildInfo
	memStats  *runtime.MemStats

	Time        time.Time
	TimeStart   time.Time
	TimeRunning time.Duration
}

func getSysInfo() *SysInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	threads, _ := runtime.ThreadCreateProfile([]runtime.StackRecord{})

	build, _ := debug.ReadBuildInfo()
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
		TimeRunning: time.Now().Sub(startTime),

		Mem:         m.Sys,
		GcPauseNs:   m.PauseTotalNs,
		LastGcPause: m.PauseNs[(m.NumGC+255)%256],
		GcCPU:       m.GCCPUFraction,

		buildInfo: build,
		memStats:  &m,
	}
}
