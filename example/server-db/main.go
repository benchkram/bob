package main

import (
	"log"
	"os"
	"os/signal"
	"server-db/server/itemrepo"
	restapi "server-db/server/rest-api"
	"syscall"

	"server-db/server"
	"server-db/server/database"
)

func main() {
	db := database.New(":6379")

	itemRepo := itemrepo.New(db)

	app := server.New(
		server.WithItemRepository(itemRepo),
	)

	restapi := restapi.New(app)

	err := restapi.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}

	// handle Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
}
