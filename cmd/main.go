package main

import (
	"log"

	"github.com/v1adhope/builder-main-droplet/internal/app"
)

func main() {
	app.CheckErr(app.Run())

	log.Print("reboot pc for apply changes")
}
