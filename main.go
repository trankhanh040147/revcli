package main

import (
	"os"

	"github.com/trankhanh040147/go-rev-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

