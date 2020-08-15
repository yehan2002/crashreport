package crashreport

import (
	"io"

	"github.com/yehan2002/crashreport/internal"
)

//Crash write the the stacktrace and profiles to to w
//Deprecated: use NewCrashReport instead
func Crash(reason string, w io.WriteCloser) {
	d := internal.CreateWriter(w)
	d.Profiles("goroutine", "heap", "allocs", "threadcreate", "block", "mutex")
	d.Reason([]string{reason})
	d.Stack()
	d.Close()
}
