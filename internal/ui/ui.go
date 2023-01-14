package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
	"github.com/yehan2002/crashreport/internal"
	"github.com/yehan2002/crashreport/internal/ui/html"
)

var exitWG sync.WaitGroup

// UI ui
type UI struct {
	port    int
	browser bool

	listener net.Listener

	server   *http.Server
	serveMux *http.ServeMux
	pages    []*page

	crashReport *internal.CrashReport

	mux sync.Mutex
	ws  []*websocket.Conn

	stop sync.Once
	run  uint32
}

type page struct {
	Name  string
	URL   template.URL
	Class string
}

var upgrader = websocket.Upgrader{EnableCompression: false}

// Run runs the web ui
func Run(r io.Reader, port int, openBrowser bool) error {
	data, err := internal.Read(r)
	if err != nil {
		return err
	}

	ui := newUI(data, port, openBrowser)

	exit, err := ui.Run()
	if err != nil {
		return err
	}
	<-exit
	return nil
}

// newUI Create a newUI ui
func newUI(data *internal.CrashReport, port int, browser bool) *UI {
	ui := &UI{port: port, browser: browser, server: &http.Server{}, crashReport: data}
	if err := ui.Init(data); err != nil {
		panic(err)
	}
	return ui
}

// serveStatic serves a static page at the given url.
func (u *UI) serveStatic(name, templateName, url string, data any) error {
	var buf bytes.Buffer
	tmp := Template.Lookup(templateName)
	if tmp == nil {
		return fmt.Errorf("template %s does not exist", templateName)
	}

	err := tmp.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error executing template %s: %w", templateName, err)
	}

	u.serveMux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(buf.Bytes())
		u.logHTTPErr(r, err)
	})

	u.pages = append(u.pages, &page{name, template.URL(url), strings.ToLower(name)})

	return nil
}

func (u *UI) logHTTPErr(r *http.Request, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serving request %s: %s", r.URL, err)
	}
}

// Init initializes the ui
func (u *UI) Init(data *internal.CrashReport) error {
	mux := http.NewServeMux()
	u.serveMux = mux

	for _, prof := range data.Profiles {
		err := prof.Register(mux)
		if err != nil {
			return err
		}

		u.pages = append(u.pages, &page{prof.Name(), template.URL("/profile/" + prof.URL()), prof.URL()})
	}

	if len(data.Stack) != 0 || len(data.Reason) != 0 {
		if err := u.serveStatic("Stack Trace", "stack.html", "/stacktrace", data); err != nil {
			return err
		}
	}

	if data.SysInfo != nil {
		if err := u.serveStatic("Info", "sys.html", "/info", data.SysInfo); err != nil {
			return err
		}
	}

	if data.Memstats != nil {
		if err := u.serveStatic("Memory", "mem.html", "/memory", data.Memstats); err != nil {
			return err
		}
	}

	reportJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	mux.HandleFunc("/report.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		_, err := w.Write(reportJSON)

		u.logHTTPErr(r, err)
	})
	u.pages = append(u.pages, &page{Name: "Report JSON", URL: "/report.json", Class: "report"})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := Template.Lookup("main.html").Execute(w, u.pages)
		u.logHTTPErr(req, err)
	})

	mux.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		u.mux.Lock()
		ws, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			u.ws = append(u.ws, ws)
		}
		u.mux.Unlock()
	})

	mux.HandleFunc("/ok", func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte("ok"))
		u.logHTTPErr(req, err)
	})

	mux.Handle("/assets/", http.FileServer(http.FS(html.Resources)))

	u.mux.Lock()
	u.serveMux = mux
	u.server.Handler = mux
	for i, ws := range u.ws {
		if ws != nil {
			err := ws.WriteMessage(websocket.TextMessage, []byte{})
			if err != nil {
				u.ws[i] = nil
			}
		}
	}
	u.mux.Unlock()

	return nil
}

// Run run the ui
func (u *UI) Run() (exit chan error, err error) {
	if !atomic.CompareAndSwapUint32(&u.run, 0, 1) {
		panic("Cannot call Run more than once")
	}

	u.listener, err = net.Listen("tcp", "localhost:"+strconv.FormatInt(int64(u.port), 10))
	if err != nil {
		return nil, err
	}

	exit = make(chan error, 1)
	address := "http://localhost:" + strings.Split(u.listener.Addr().String(), ":")[1]
	fmt.Println("Serving web UI on " + address)

	if u.browser {
		go func() { time.Sleep(time.Second * 2); browser.OpenURL(address) }()
	}

	go func() { exitWG.Wait(); u.Stop() }()
	go func() { exit <- u.server.Serve(u.listener) }()

	return
}

// Exit returns when the server exits
func (u *UI) Exit() error {
	return nil
}

// Stop stop the server
func (u *UI) Stop() (err error) {
	u.stop.Do(func() {
		err = u.server.Close()
	})
	return
}

func init() {
	exitWG.Add(1)
	go func() {
		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt)
		<-sig
		exitWG.Done()
	}()
}
