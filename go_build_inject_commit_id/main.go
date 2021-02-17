package main

import (
	"flag"
	"fmt"
)

var (
	VersionCmd   bool
	BuildVersion = "unknown"
	BuildTime    = "unknown"
)

func main() {
	flag.BoolVar(&VersionCmd, "v", false, "show version")
	flag.Parse()

	if VersionCmd {
		fmt.Printf("build time: %s\nbuild commit_id: %s\n", BuildTime, BuildVersion)

		return
	}

	fmt.Println("hello world")
}
