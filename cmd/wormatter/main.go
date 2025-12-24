package main

import (
	"fmt"
	"os"

	"github.com/werf/wormatter/pkg/formatter"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: wormatter <file.go|directory>")
		os.Exit(1)
	}

	path := os.Args[1]
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if info.IsDir() {
		err = formatter.FormatDirectory(path)
	} else {
		err = formatter.FormatFile(path)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
