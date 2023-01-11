package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yehan2002/crashreport/internal/ui"
	"golang.org/x/term"
)

func main() {
	var port uint
	var openBrowser bool
	flag.UintVar(&port, "port", 0, "The port to use")
	flag.BoolVar(&openBrowser, "browser", !term.IsTerminal(int(os.Stdout.Fd())), "Open the result in a browser")
	flag.Parse()

	fileName := flag.Arg(0)
	if len(fileName) == 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "File was not specified")
		flag.Usage()
		return
	}

	if flag.NArg() > 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Expected exactly one argument got %d", flag.NArg())
		flag.Usage()
		return
	}

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Unable to open %s: %s", fileName, err)
		return
	}
	defer file.Close()

	err = ui.Run(file, int(port), openBrowser)
	if err != nil {
		fmt.Fprint(flag.CommandLine.Output(), err.Error())
	}
}
