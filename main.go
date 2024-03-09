package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Mat12143/TCPGame/modules"
)

func main() {

	arg := os.Args

	switch arg[1] {

	case "1":
		{
			server := modules.NewServer("localhost", "3000")
			log.Fatal(server.Start())
		}
	case "2":
		{
			var username string

			fmt.Print("Enter username: ")
			fmt.Scanf("%s", &username)

			client := modules.NewClient("localhost:3000", username)
			log.Fatal(client.StartClient())

		}
	}

}
