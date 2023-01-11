package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yehan2002/crashreport/internal/ui"
	"golang.org/x/term"
)

func main() {
	var port uint
	var openBrowser bool
	flag.UintVar(&port, "port", 0, "The port to use. Defaults to using a random port.")
	flag.BoolVar(&openBrowser, "browser", !term.IsTerminal(int(os.Stdout.Fd())), "Opens the crash report in a browser.")

	var printFullHelpMessage = true
	flag.Usage = func() {
		bin := filepath.Base(os.Args[0])
		out := flag.CommandLine.Output()
		if printFullHelpMessage {
			fmt.Fprintf(out, "crashreport is a tool for viewing crash reports:\n\n")
		}
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "\t%s [OPTION]... FILE\n\n", bin)
		fmt.Fprintf(out, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	printFullHelpMessage = false

	fileName := flag.Arg(0)
	if len(fileName) == 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "File was not specified\n\n")
		flag.Usage()
		return
	}

	if flag.NArg() > 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Expected exactly one argument got %d.\n\n", flag.NArg())
		flag.Usage()
		return
	}

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Unable to open %s: %s\n", fileName, err)
		return
	}
	defer file.Close()

	err = ui.Run(file, int(port), openBrowser)
	if err != nil {
		fmt.Fprintln(flag.CommandLine.Output(), err.Error())
	}
}
