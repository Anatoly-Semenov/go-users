package main

import (
	"log"

	"github.com/anatoly_dev/go-users/cmd/app/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}
