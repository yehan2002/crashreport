package internal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

// maxSize the max size for a file inside the crash report
const maxSize = 1024 * 1024 // 1MB

// Read reads a crash report from the zip file
func Read(r io.Reader) (report *CrashReport, err error) {
	report = &CrashReport{
		Build:    &debug.BuildInfo{},
		SysInfo:  &SysInfo{},
		Memstats: &runtime.MemStats{},
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return nil, fmt.Errorf("unable read zip file: %w", err)
	}

	if err = report.readToString(zr, "reason", &report.Reason); err != nil {
		return nil, err
	}

	if err = report.readToString(zr, "stack", (*string)(&report.Stack)); err != nil {
		return nil, err
	}

	if err = report.readJSON(zr, "build.json", &report.Build); err != nil {
		return nil, err
	}

	if err = report.readJSON(zr, "system.json", &report.SysInfo); err != nil {
		return nil, err
	}

	if err = report.readJSON(zr, "memstats.json", &report.Memstats); err != nil {
		return nil, err
	}

	if report.Files, err = fs.Glob(zr, "include/*"); err != nil {
		return nil, fmt.Errorf("Unable to get list of included files: %w", err)
	}

	// read all profiles in the zip file.
	if err = report.readProfiles(zr); err != nil {
		return nil, err
	}

	return report, nil
}

// readProfiles reads all profile files in the given fs
func (c *CrashReport) readProfiles(f fs.FS) error {
	profiles, err := fs.Glob(f, "profiles/*.prof")
	if err != nil {
		return fmt.Errorf("unable to find profile files: %w", err)
	}

	for _, profileName := range profiles {
		buf, err := c.readFile(f, profileName)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(path.Base(profileName), ".prof")
		c.Profiles = append(c.Profiles, NewProfile(name, buf))
	}

	return nil
}

// readJSON reads and parses the given file into dst.
// dst must be a non nil pointer to a pointer to struct (**struct)
func (c *CrashReport) readJSON(f fs.FS, name string, dst any) error {
	v := reflect.ValueOf(dst)

	buf, err := c.readFile(f, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			v.Elem().Set(reflect.Zero(v.Elem().Type()))
			return nil
		}
		return err
	}

	err = json.Unmarshal(buf, v.Elem().Interface())
	if err != nil {
		return err
	}

	return nil
}

// readToString reads the given file into dst.
func (c *CrashReport) readToString(f fs.FS, name string, dst *string) error {
	buf, err := c.readFile(f, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	*dst = string(buf)
	return nil
}

func (c *CrashReport) readFile(f fs.FS, name string) (buf []byte, err error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", name, err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error calling stat on %s: %w", name, err)
	}

	if size := stat.Size(); size > maxSize {
		return nil, fmt.Errorf("file %s exceeds max size: size %d, max: %d", name, size, maxSize)
	}

	buf, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", name, err)
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return
}
