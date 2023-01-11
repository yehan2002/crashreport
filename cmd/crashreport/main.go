package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/yehan2002/crashreport/internal"
	"github.com/yehan2002/crashreport/internal/ui"
	"golang.org/x/term"
)

func main() {
	var port uint
	var openBrowser bool
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.CommandLine.PrintDefaults()
	}

	flag.UintVar(&port, "port", 0, "The port to use")
	flag.BoolVar(&openBrowser, "browser", term.IsTerminal(int(os.Stdout.Fd())), "Open the result in a browser")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("No files specified\n")
		flag.Usage()
	} else if len(args) > 1 {
		fmt.Printf("Expected only one file but got %d\n", len(args))
		flag.Usage()
	}
	run(args[0], int(port), openBrowser)
}

func run(file string, port int, openBrowser bool) {
	data, err := internal.Read(file)
	panicErr(err)

	ui := ui.New(data, port, openBrowser)

	watcher, err := fsnotify.NewWatcher()
	panicErr(err)
	panicErr(watcher.Add(file))

	go func() { panicErr(<-watcher.Errors) }()
	go func() {
		for range watcher.Events {
			data, err := internal.Read(file)
			if err == nil {
				ui.Reload(data)
			}
		}
	}()

	exit, err := ui.Run()
	panicErr(err)
	<-exit
	panicErr(watcher.Close())
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
