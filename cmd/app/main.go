package main

import (
	"github.com/dronesphere/configs"
	"github.com/dronesphere/internal/app"
)

func main() {
	// TODO: Load configuration
	cfg := configs.Config{}

	// Run
	app.Run(&cfg)
}
