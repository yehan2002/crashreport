package crashreport

import (
	"archive/zip"
	"bytes"
	"io"
	"runtime"
	"runtime/pprof"
)

//Crash crash the program and write dump files
func Crash(reason string, w io.WriteCloser) {
	d := crashWriter{zip.NewWriter(w)}
	d.Profile("goroutine", "heap", "allocs", "threadcreate", "block", "mutex")
	d.Stack(reason)
	d.Close()
}

type crashWriter struct {
	archive *zip.Writer
}

func (d *crashWriter) panic(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

//Profile write the provided profiles to the crash file
func (d *crashWriter) Profile(names ...string) {
	for _, name := range names {
		w := d.open(name + ".prof")
		d.panic(nil, pprof.Lookup(name).WriteTo(w, 0))
	}
}

//Stack write the complete stack to the crash file
func (d *crashWriter) Stack(reason string) {
	w := d.open("stack")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	d.panic(w.Write([]byte(reason)))
	d.panic(io.CopyN(w, bytes.NewBuffer(buf), int64(n)))
}

//Open open a file inside the crash file
func (d *crashWriter) open(name string) io.Writer {
	return d.panic(d.archive.Create(name)).(io.Writer)
}

//Close close the crashfile
func (d *crashWriter) Close() {
	d.panic(nil, d.archive.Close())
}
