package main

import (
	"github.com/dronesphere/configs"
	"github.com/dronesphere/internal/app"
)

func main() {
	// if err := godotenv.Load(); err != nil {
	// 	panic("Error loading .env file")
	// }
	cfg, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Run
	app.Run(cfg)
}
