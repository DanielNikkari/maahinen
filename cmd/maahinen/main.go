package main

import (
	"fmt"
	"os"

	"maahinen/internal/setup"
)

func main() {
	if err := setup.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
