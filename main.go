package main

import (
	"log"
	"myproject/config"
)

func main() {
	router := config.SetupRouter()
	err := router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
