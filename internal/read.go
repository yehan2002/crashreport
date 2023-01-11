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
	"strings"
)

// maxSize the max size for a file inside the crash report
const maxSize = 1024 * 1024 // 1MB

// Read reads a crash report from the zip file
func Read(filename string) (d *CrashReport, err error) {
	data := &CrashReport{}

	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %w", filename, err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return nil, fmt.Errorf("unable read zip file %s: %w", filename, err)
	}

	if err = data.readToString(zr, "reason", &d.Reason); err != nil {
		return nil, err
	}

	if err = data.readToString(zr, "stack", &d.Stack); err != nil {
		return nil, err
	}

	if err = data.readJSON(zr, "build.json", &data.Build); err != nil {
		return nil, err
	}

	if err = data.readJSON(zr, "system.json", &data.SysInfo); err != nil {
		return nil, err
	}

	if err = data.readJSON(zr, "memstats.json", &data.Memstats); err != nil {
		return nil, err
	}

	// read all profiles in the zip file.
	if err = data.readProfiles(zr); err != nil {
		return nil, err
	}

	return data, nil
}

// readProfiles reads all profile files in the given fs
func (c *CrashReport) readProfiles(f fs.FS) error {
	profiles, err := fs.Glob(f, "*.prof")
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

func (c *CrashReport) readJSON(f fs.FS, name string, dst interface{}) error {
	buf, err := c.readFile(f, name)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, err)
	if err != nil {
		return err
	}

	return nil
}

func (c *CrashReport) readToString(f fs.FS, name string, dst *string) error {
	buf, err := c.readFile(f, name)
	if err != nil {
		return err
	}

	*dst = string(buf)
	return nil
}

func (c *CrashReport) readFile(f fs.FS, name string) (buf []byte, err error) {
	file, err := f.Open(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
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
