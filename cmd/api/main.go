package main

import (
	"log"

	"github.com/muktiarafi/ticketing-orders/internal/server"
)

func main() {
	e := server.SetupServer()

	log.Fatal(e.Start(":8080"))
}
