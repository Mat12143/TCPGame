package main

import (
	"log"

	"github.com/Mat12143/TCPGame/server/modules"
)

func main() {
	server := modules.NewServer("localhost", "3000")
	log.Fatal(server.Start())
}
