package internal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

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
	Reason []string

	NoStack   bool
	NoSysInfo bool

	Profiles map[string]struct{}
	Files    []string
}

func Create(c Config) (*CrashReport, error) {
	cr := CrashReport{
		Reason: strings.Join(c.Reason, "\n"),
		Files:  c.Files,
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	cr.Memstats = &mem

	if !c.NoStack {
		buf := make([]byte, 1<<16)
		n := runtime.Stack(buf, true)
		cr.Stack = string(buf[:n])
	}

	if !c.NoSysInfo {
		cr.SysInfo = newSysInfo()
	}

	for profile := range c.Profiles {
		prof := pprof.Lookup(profile)
		if prof == nil {
			return nil, fmt.Errorf("unable to find profile %s", profile)
		}

		var buf bytes.Buffer
		err := prof.WriteTo(&buf, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to write profile %s: %w", profile, err)
		}
		cr.Profiles = append(cr.Profiles, NewProfile(profile, buf.Bytes()))
	}

	return &cr, nil
}

func (c *CrashReport) Write(w io.Writer) error {
	zw := zip.NewWriter(w)
	n, err := w.Write([]byte(Header))
	if err != nil {
		return fmt.Errorf("unable to write header: %w", err)
	}
	zw.SetOffset(int64(n))

	if err = c.writeJSON(zw, "build.json", c.Build); err != nil {
		return err
	}
	if err = c.writeJSON(zw, "memstats.json", c.Memstats); err != nil {
		return err
	}
	if err = c.writeJSON(zw, "system.json", c.SysInfo); err != nil {
		return err
	}
	if err = c.write(zw, "reason", strings.NewReader(c.Reason)); err != nil {
		return err
	}
	if err = c.write(zw, "stack", strings.NewReader(c.Stack)); err != nil {
		return err
	}

	for _, profile := range c.Profiles {
		if err = c.write(zw, profile.Name()+".prof", bytes.NewReader(profile.profile)); err != nil {
			return err
		}
	}

	for _, file := range c.Files {
		if err = c.writeFile(zw, file); err != nil {
			return err
		}
	}

	return zw.Close()
}

func (c *CrashReport) writeJSON(z *zip.Writer, name string, v interface{}) error {
	if v == nil {
		return nil
	}

	w, err := z.Create(name)
	if err != nil {
		return fmt.Errorf("unable to create file %s in zip archive: %w", name, err)
	}

	err = json.NewEncoder(w).Encode(v)
	if err != nil {
		return fmt.Errorf("error while writing json file %s: %w", name, err)
	}
	return nil
}

func (c *CrashReport) write(z *zip.Writer, name string, data io.Reader) error {
	w, err := z.Create(name)
	if err != nil {
		return fmt.Errorf("unable to create file %s in zip archive: %w", name, err)
	}

	_, err = io.Copy(w, data)
	if err != nil {
		return fmt.Errorf("error while writing file %s: %w", name, err)
	}

	return nil
}

func (c *CrashReport) writeFile(zw *zip.Writer, file string) error {
	var err error
	if file, err = filepath.Abs(file); err != nil {
		return err
	}

	name := filepath.Base(file)

	var f *os.File
	if f, err = os.Open(file); err != nil {
		return err
	}

	w, err := zw.Create("include/" + name)
	if err != nil {
		return err
	}

	if _, err = io.Copy(w, f); err != nil {
		return err
	}

	return f.Close()
}
