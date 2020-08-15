package ui

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/pprof/driver"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
	"github.com/yehan2002/crashreport/internal"
	"github.com/yehan2002/crashreport/internal/ui/html"
)

var exitWG sync.WaitGroup

//UI ui
type UI struct {
	port    int
	browser bool

	listener net.Listener

	serverMux sync.Mutex
	server    *http.Server
	ws        []*websocket.Conn

	stop sync.Once
	run  uint32
}

type page struct {
	Name  string
	URL   template.URL
	Class string
}

var upgrader = websocket.Upgrader{EnableCompression: false}

//New Create a new ui
func New(data *internal.Data, port int, browser bool) *UI {
	ui := &UI{port: port, browser: browser, server: &http.Server{}}
	ui.Reload(data)
	return ui
}

//Reload reload the page
func (u *UI) Reload(data *internal.Data) {
	mux := http.NewServeMux()
	pages := []*page{}
	for _, prof := range data.Profiles {
		driver.PProf(&driver.Options{HTTPServer: func(d *driver.HTTPServerArgs) error {
			for path, handler := range d.Handlers {
				mux.Handle("/profile/"+prof.URL+path, handler)
			}
			return nil
		}, UI: &profUI{}, Flagset: &fakeFlags{}, Fetch: prof})
		pages = append(pages, &page{prof.Name, template.URL("/profile/" + prof.URL), prof.URL})
	}

	if len(data.Stack) != 0 || len(data.Reason) != 0 {
		mux.HandleFunc("/stacktrace", func(w http.ResponseWriter, _ *http.Request) {
			html.Template.Lookup("stack.html").Execute(w, data)
		})
		pages = append(pages, &page{"Stack Trace", template.URL("/stacktrace"), "stacktrace"})
	}

	if data.SysInfo != nil {
		mux.HandleFunc("/info", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Println(html.Template.Lookup("sys.html").Execute(w, data.SysInfo))
		})
		pages = append(pages, &page{"Info", template.URL("/info"), "info"})
	}

	if data.Memstat != nil {
		mux.HandleFunc("/memory", func(w http.ResponseWriter, _ *http.Request) {
			fmt.Println(html.Template.Lookup("mem.html").Execute(w, data.Memstat))
		})
		pages = append(pages, &page{"Memory", template.URL("/memory"), "memory"})
	}

	mux.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		u.serverMux.Lock()
		ws, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			u.ws = append(u.ws, ws)
		}
		u.serverMux.Unlock()
	})

	mux.HandleFunc("/ok", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			w.WriteHeader(404)
			w.Write([]byte("Not Found"))
		} else {
			html.Template.Lookup("main.html").Execute(w, pages)
		}
	})

	u.serverMux.Lock()
	u.server.Handler = mux
	for i, ws := range u.ws {
		if ws != nil {
			err := ws.WriteMessage(websocket.TextMessage, []byte{})
			if err != nil {
				u.ws[i] = nil
			}
		}
	}
	u.serverMux.Unlock()
}

//Run run the ui
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

//Exit returns when the server exits
func (u *UI) Exit() error {
	return nil
}

//Stop stop the server
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
