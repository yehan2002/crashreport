package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/yehan2002/crashreport/internal"
	"github.com/yehan2002/crashreport/internal/ui"
)

func main() {
	var port *int
	var openBrowser *bool
	cmd := cobra.Command{
		Use:   "crashreport",
		Short: "View crash reports",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("No files specified")
			} else if len(args) > 1 {
				return fmt.Errorf("Expected only one file but got %d", len(args))
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			run(args[0], *port, *openBrowser)
		},
	}
	port = cmd.Flags().IntP("port", "p", 0, "The port to use")
	openBrowser = cmd.Flags().BoolP("browser", "b", false, "Open the result in a browser")
	cmd.Execute()
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
		for v := range watcher.Events {
			data, err := internal.Read(file)
			if err == nil {
				ui.Reload(data)
				fmt.Println(v)
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
