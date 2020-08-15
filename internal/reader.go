package internal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/pprof/profile"
)

//maxSize the max size for a file inside the crashreport
const maxSize = 128 * 1024 // 128kb

//Data data
type Data struct {
	Profiles []*Profile
	SysInfo  *SysInfo
	Memstat  *runtime.MemStats
	Build    *debug.BuildInfo

	Reason string
	Stack  string
}

//Profile a profile
type Profile struct {
	Profile *profile.Profile
	Name    string
	URL     string
}

//Fetch used by pprof to read this
func (p *Profile) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	return p.Profile, "", nil
}

//Read read the zip file
func Read(filename string) (d *Data, err error) {
	data := &Data{}
	var buf []byte
	var zr *zip.Reader

	if buf, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	if zr, err = zip.NewReader(bytes.NewReader(buf), int64(len(buf))); err != nil {
		return
	}

	for _, file := range zr.File {
		var buf []byte

		if strings.HasSuffix(file.Name, ".prof") {
			var prof *profile.Profile
			name := strings.TrimSuffix(path.Base(file.Name), ".prof")
			if buf, err = readFile(file); err == nil {
				if prof, err = profile.ParseData(buf); err == nil {
					data.Profiles = append(data.Profiles, &Profile{Profile: prof, URL: name, Name: strings.Title(name)})
				}
				continue
			}
		}

		switch file.Name {
		case "build.json":
			var t debug.BuildInfo
			if buf, err = readFile(file); err == nil {
				if err = json.Unmarshal(buf, &t); err == nil {
					data.Build = &t
					continue
				}
			}
		case "system.json":
			var t SysInfo
			if buf, err = readFile(file); err == nil {
				if err = json.Unmarshal(buf, &t); err == nil {
					data.SysInfo = &t
					continue
				}
			}

		case "memstats.json":
			var t runtime.MemStats
			if buf, err = readFile(file); err == nil {
				if err = json.Unmarshal(buf, &t); err == nil {
					data.Memstat = &t
					continue
				}
			}

		case "reason":
			if buf, err = readFile(file); err == nil {
				data.Reason = string(buf)
				continue
			}

		case "stack":
			if buf, err = readFile(file); err == nil {
				data.Stack = string(buf)
				continue
			}
		default:
			continue
		}
		return nil, err
	}

	return data, nil
}

func readFile(file *zip.File) (data []byte, err error) {
	var r io.ReadCloser
	var tmp []byte

	if file.UncompressedSize64 > maxSize {
		return nil, fmt.Errorf("File \"%s\" exceeds max size", file.Name)
	}

	if r, err = file.Open(); err != nil {
		return
	}
	if tmp, err = ioutil.ReadAll(r); err != nil {
		return
	}
	if err = r.Close(); err != nil {
		return
	}
	return tmp, nil
}
