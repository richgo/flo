package main

import (
	"fmt"
	"os"

	"github.com/richgo/enterprise-ai-sdlc/cmd/eas/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
