package main

import (
	"finance-app-go-demo/app"
	"finance-app-go-demo/config"

	_ "finance-app-go-demo/docs"
)

// @title Finance App API
// @version 1.0
// @description Finance application API
// @host localhost:3100
// @BasePath /api
func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	err = app.Run(cfg)
	if err != nil {
		panic(err)
	}
}
