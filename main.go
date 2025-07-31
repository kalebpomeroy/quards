package main

import (
	"log"
	app "quards/internal/api"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
