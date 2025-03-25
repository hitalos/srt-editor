package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hitalos/srt-editor/application"
)

var (
	gitCommit = "DEV"
	version   = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println("Build version:", gitCommit)
		os.Exit(0)
	}

	inputFile := flag.Arg(0)
	if inputFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	srt := new(application.Srt)
	if err := srt.Load(inputFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := application.New(srt)

	if err := app.Run(); err != nil {
		panic(err)
	}

}
