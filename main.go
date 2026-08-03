package main

import (
	"fmt"
	"os"

	"github.com/m-tkg/md2site/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "md2site:", err)
		os.Exit(1)
	}
}
