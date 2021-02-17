package main

import (
	"flag"
	"fmt"
)

var cmd = flag.String("cmd", "read", "read data")
var path = flag.String("path", ".", "the path of data")
var a2i = flag.Bool("a2i", false, "enable a2i")

func main() {

	flag.Parse()
	switch *cmd {
	case "read":
		fmt.Printf("nameWithDra or NameWithoutDra\n")
	default:

		fmt.Printf("default\n")
	}

}
