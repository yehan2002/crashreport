package internal

import (
	"archive/zip"
	"bytes"
	"io"
	"runtime"
	"runtime/pprof"
)

//CrashWriter a writer for writing crashes
type CrashWriter struct {
	Archive *zip.Writer
}

func (d *CrashWriter) panic(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

//Profile write the provided profiles to the crash file
func (d *CrashWriter) Profile(names ...string) {
	for _, name := range names {
		w := d.open(name + ".prof")
		d.panic(nil, pprof.Lookup(name).WriteTo(w, 0))
	}
}

//Stack write the complete stack to the crash file
func (d *CrashWriter) Stack(reason string) {
	w := d.open("stack")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	d.panic(w.Write([]byte(reason)))
	d.panic(io.CopyN(w, bytes.NewBuffer(buf), int64(n)))
}

//Open open a file inside the crash file
func (d *CrashWriter) open(name string) io.Writer {
	return d.panic(d.Archive.Create(name)).(io.Writer)
}

//Close close the crashfile
func (d *CrashWriter) Close() {
	d.panic(nil, d.Archive.Close())
}
