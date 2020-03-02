package main

import (
	"flag"
	"log"

	"github.com/yehan2002/crashreport/internal/ui"
)

func main() {
	var port = flag.Int("port", 0, "The port to use")
	var openBrowser = flag.Bool("browser", false, "Open the result in a browser")

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Please provide one crash log")
	}
	ui := ui.WebUI{Port: int64(*port), Browser: *openBrowser}
	panicErr(ui.LoadZip(args[0]))
	ui.Show()
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
