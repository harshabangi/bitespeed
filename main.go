package main

import (
	"github.com/harshabangi/bitespeed/internal/service"
	"log"
)

func main() {
	app, err := service.NewService()
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
