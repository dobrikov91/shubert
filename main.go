package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"strings"
)

//go:embed version.txt
var versionRaw string

var version = strings.TrimSpace(versionRaw)

func main() {
	port := flag.String("port", "8088", "web port")
	debug := flag.Bool("kb", false, "mock midi input with laptop keyboard")
	versionFlag := flag.Bool("version", false, "print version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", version)
		return
	}

	c, err := NewController("./data/config.json", *port, version)
	if err != nil {
		log.Fatalf("Can't create controller %v", err)
	}

	w := NewWebserver(c)
	c.web = w

	if *debug {
		go MidiMock(c)
	}
	go c.MainLoop(*debug)

	w.Run()
}
