package crashreport

import (
	"io"

	"github.com/yehan2002/crashreport/internal"
)

// Crash write the the stacktrace and profiles to to w
//
// Deprecated: This panics if any errors occur. Use NewCrashReport instead.
func Crash(reason string, w io.WriteCloser) {
	d, err := internal.Create(internal.Config{
		Reason: []string{reason},
		Profiles: map[string]struct{}{
			"goroutine": {}, "heap": {}, "allocs": {}, "threadcreate": {}, "block": {}, "mutex": {},
		},
	})
	if err != nil {
		panic(err)
	}

	if err := d.Write(w); err != nil {
		panic(err)
	}

	if err := w.Close(); err != nil {
		panic(err)
	}
}
