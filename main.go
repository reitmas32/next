package main

import (
	"os"

	"github.com/reitmas32/next/cmd/next"
)

func main() {
	if err := next.Execute(); err != nil {
		os.Exit(1)
	}
}
