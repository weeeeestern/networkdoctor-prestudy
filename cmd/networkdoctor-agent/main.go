//go:build linux

package main

import (
	"log"

	"networkdoctor-agent/internal/app"
	"networkdoctor-agent/internal/config"
)

func main() {
	cfg := config.ParseFlags()
	if err := app.Run(cfg); err != nil {
		log.Fatalf("networkdoctor agent failed: %v", err)
	}
}
