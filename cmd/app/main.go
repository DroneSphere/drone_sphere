package main

import (
	"github.com/dronesphere/configs"
	"github.com/dronesphere/internal/app"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Run
	app.Run(cfg)
}
