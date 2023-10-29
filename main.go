package main

import (
	"fmt"
	"os"

	"github.com/simontheleg/konf-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "konf execution has failed: %q\n", err)
		os.Exit(1)
	}
}
