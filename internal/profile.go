package internal

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
)

// Profile a profile
type Profile struct {
	profile []byte
	name    string
}

func (p *Profile) URL() string  { return strings.ToLower(p.name) }
func (p *Profile) Name() string { return p.name }

func (p *Profile) Profile() (*profile.Profile, error) {
	return profile.ParseData(p.profile)
}

func (p *Profile) ProfileBytes() []byte { return p.profile }

func (p *Profile) Register(mux *http.ServeMux) error {
	prof, err := p.Profile()
	if err != nil {
		return err
	}

	return driver.PProf(&driver.Options{HTTPServer: func(d *driver.HTTPServerArgs) error {
		for path, handler := range d.Handlers {
			u, err := url.JoinPath("/profile", p.URL(), path)
			if err != nil {
				return err
			}

			mux.Handle(u, handler)
		}
		return nil
	}, UI: &profUI{}, Flagset: &fakeFlags{}, Fetch: &fetcher{P: prof}})
}

func NewProfile(name string, prof []byte) *Profile {
	return &Profile{profile: prof, name: strings.Title(name)}
}
