package main

import (
	"os"

	"github.com/trankhanh040147/revcli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
