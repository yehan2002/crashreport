package crashreport

import (
	"archive/zip"
	"io"

	"github.com/yehan2002/crashreport/internal"
)

//Crash write the the stacktrace and profiles to to w
func Crash(reason string, w io.WriteCloser) {
	d := internal.CrashWriter{Archive: zip.NewWriter(w)}
	d.Profile("goroutine", "heap", "allocs", "threadcreate", "block", "mutex")
	d.Stack(reason)
	d.Close()
}
