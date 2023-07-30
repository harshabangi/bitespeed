package main

import (
	"github.com/harshabangi/bitespeed/internal/service"
	"log"
)

// @title BiteSpeed API
// @version 1.0
// @description BiteSpeed Server
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	app, err := service.NewService()
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
