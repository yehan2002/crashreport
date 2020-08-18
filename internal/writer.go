package internal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

//Header the header line to be used for the file.
var Header = `crashreport
Use github.com/yehan2002/crashreport or open this file with any zip file viewer.
`

//CrashWriter a writer for writing crashes
type CrashWriter struct {
	archive *zip.Writer
}

//CreateWriter create a new writer
func CreateWriter(w io.Writer) *CrashWriter {
	c := &CrashWriter{archive: zip.NewWriter(w)}
	n := c.Panic(w.Write([]byte(Header))).(int)
	c.archive.SetOffset(int64(n))
	return c
}

//Panic panic if err != nil
func (d *CrashWriter) Panic(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}

//Profiles write the provided profiles to the crash file
func (d *CrashWriter) Profiles(names ...string) {
	for _, name := range names {
		d.Profile(name)
	}
}

//File includes the given file in the crash report
func (d *CrashWriter) File(file string) {
	var err error
	if file, err = filepath.Abs(file); err != nil {
		return
	}

	name := filepath.Base(file)

	if name == "" || (len(name) == 1 && name[0] == os.PathSeparator) {
		return //unreachable
	}

	var f *os.File
	if f, err = os.Open(file); err != nil {
		return
	}
	w := d.open("include/" + file)
	io.Copy(w, f)
	f.Close()
}

//Profile write the given profile to the crash file
func (d *CrashWriter) Profile(name string) {
	w := d.open("profiles/" + name + ".prof")
	d.Panic(nil, pprof.Lookup(name).WriteTo(w, 0))
}

//Stack write the complete stack to the crash file
func (d *CrashWriter) Stack() {
	w := d.open("stack")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)

	d.Panic(io.CopyN(w, bytes.NewBuffer(buf), int64(n)))

}

//Reason write the reason to the crash report
func (d *CrashWriter) Reason(r []string) {
	w := d.open("reason")
	for _, line := range r {
		d.Panic(w.Write([]byte(line)))
		d.Panic(w.Write([]byte{'\n'}))
	}
}

//SysInfo write system info to the crash report
func (d *CrashWriter) SysInfo() {
	sys := getSysInfo()

	w := d.open("system.json")
	buf := d.Panic(json.Marshal(sys)).([]byte)
	d.Panic(w.Write(buf))

	w = d.open("memstats.json")
	buf = d.Panic(json.Marshal(sys.memStats)).([]byte)
	d.Panic(w.Write(buf))

	if sys.buildInfo != nil {
		w = d.open("build.json")
		buf = d.Panic(json.Marshal(sys.buildInfo)).([]byte)
		d.Panic(w.Write(buf))
	}
}

//Open open a file inside the crash file
func (d *CrashWriter) open(name string) io.Writer {
	return d.Panic(d.archive.Create(name)).(io.Writer)
}

//Close close the crashfile
func (d *CrashWriter) Close() {
	d.Panic(nil, d.archive.Close())
}
