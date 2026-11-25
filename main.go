package main

import (
	"fmt"
	"os"

	"k8s-assistant/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
