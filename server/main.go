package main

import (
	"log"

	"github.com/Mat12143/TCPGame/server/modules"
)

func main() {
	server := modules.NewServer("0.0.0.0", "3000")
	log.Fatal(server.Start())
}
