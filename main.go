package main

import (
	"os"

	"github.com/rafa/next/cmd/next"
)

func main() {
	if err := next.Execute(); err != nil {
		os.Exit(1)
	}
}
