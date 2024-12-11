package main

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app, err := AppFromEnvironment()
	if err != nil {
		log.Fatalf("Unable to initialize application: %v", err.Error())
	}

	app.Run()
}
