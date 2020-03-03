package ui

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/pprof/driver"
	"github.com/pkg/browser"
	"github.com/yehan2002/crashreport/internal"
	"github.com/yehan2002/crashreport/internal/ui/html"
)

//WebUI ui
type WebUI struct {
	mux   *http.ServeMux
	pages []*page
	err   error

	tmp string

	Port    int64
	Browser bool

	listener net.Listener
	server   *http.Server
	exit     chan bool
}

type page struct {
	Name string
	URL  template.URL
}

//RegisterPProf register a new pprof profile
func (p *WebUI) RegisterPProf(fileName, url, name string) error {
	if p.mux == nil {
		p.mux = http.NewServeMux()
	}

	//TODO: find a better way to do this
	driver.PProf(&driver.Options{HTTPServer: func(d *driver.HTTPServerArgs) error {
		for path, handler := range d.Handlers {
			p.mux.Handle("/"+url+path, handler)
		}
		return nil
	}, Flagset: internal.NewFakeFlag(fileName), UI: &ui{p}})

	p.pages = append(p.pages, &page{name, template.URL("/" + url)})
	return p.err
}

//RegisterStack register a stack trace
func (p *WebUI) RegisterStack(filename, url, name string) error {
	if p.mux == nil {
		p.mux = http.NewServeMux()
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var text = string(data)
	p.mux.HandleFunc("/"+url, func(w http.ResponseWriter, _ *http.Request) {
		html.Stack.Execute(w, text)
	})

	p.pages = append(p.pages, &page{name, template.URL("/" + url)})
	return nil
}

//LoadZip load a zip containing profile
func (p *WebUI) LoadZip(file string) error {
	var err error
	p.tmp, err = ioutil.TempDir("", "go-crash")
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	p.exit = make(chan bool, 1)
	go p.shutdown()

	r, err := zip.OpenReader(file)
	if err != nil {
		return err
	}

	for _, file := range r.File {
		name, url, ispprof, ok := p.guessType(file.Name)
		if !ok {
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return err
		}
		filename := path.Join(p.tmp, path.Base(file.Name))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, fr)
		if err != nil {
			return err
		}

		if ispprof {
			err = p.RegisterPProf(filename, url, name)
			if err != nil {
				return err
			}
			continue
		}

		err = p.RegisterStack(filename, url, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *WebUI) guessType(filename string) (name, url string, ispprof, ok bool) {
	switch path.Base(filename) {
	case "allocs.prof":
		name, url, ispprof, ok = "Allocations", "allocs", true, true
	case "block.prof":
		name, url, ispprof, ok = "Block", "block", true, true
	case "goroutine.prof":
		name, url, ispprof, ok = "Goroutines", "goroutine", true, true
	case "heap.prof":
		name, url, ispprof, ok = "Heap", "heap", true, true
	case "mutex.prof":
		name, url, ispprof, ok = "Mutex", "mutex", true, true
	case "threadcreate.prof":
		name, url, ispprof, ok = "Threads", "threadcreate", true, true
	case "stack":
		name, url, ispprof, ok = "Stack", "stack", false, true
	default:
		fmt.Println("Skipping " + filename)
	}
	return
}

//Show show the web ui at a random port
func (p *WebUI) Show() error {
	var err error
	p.listener, err = net.Listen("tcp", "localhost:"+strconv.FormatInt(p.Port, 10))
	if err != nil {
		return err
	}
	p.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { html.Main.Execute(w, p.pages) })

	port := strings.Split(p.listener.Addr().String(), ":")[1]
	fmt.Println("Serving web UI on http://localhost:" + port)

	if p.Browser {
		go p.browser(port)
	}

	p.server = &http.Server{Handler: p.mux}
	err = p.server.Serve(p.listener)
	if err == http.ErrServerClosed {
		<-p.exit
	}
	return err
}

func (p *WebUI) browser(port string) {
	time.Sleep(time.Second * 2)
	browser.OpenURL("http://localhost:" + port)

}

func (p *WebUI) shutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	<-c
	fmt.Println("Shutting down server....")
	if p.listener != nil {
		p.listener.Close()
	}
	if p.server != nil {
		p.server.Close()
	}
	os.RemoveAll(p.tmp)
	p.exit <- true
}
