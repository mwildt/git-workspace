package main

import (
	"fmt"
	"os"
)

func main() {
	token := os.Getenv("GIT_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "GIT_TOKEN environment variable is required\n")
		os.Exit(1)
	}
	fmt.Print(token)
}
