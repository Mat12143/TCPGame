package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type ServerData struct {
	Status   string    `json:"status"`
	Zones    int       `json:"zones"`
	Winner   string    `json:"winner"`
	Message  string    `json:"message"`
	Upgrades []Upgrade `json:"upgrades"`
}

type Upgrade struct {
	Defence int `json:"defence"`
	Attack  int `json:"attack"`
}

type Client struct {
	address  string
	username string
}

var zones map[int]string
var zoneUpgrades []Upgrade

func parseZones(sd ServerData) {
	zones = make(map[int]string)

	for i := 0; i < sd.Zones; i++ {
		zones[i] = fmt.Sprintf("Zone %d", i)
	}
}

func NewClient(address, username string) *Client {
	return &Client{
		address:  address,
		username: username,
	}
}

func (c *Client) StartClient() error {

	ip, err := net.ResolveTCPAddr("tcp", c.address)

	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, ip)

	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(c.username))

	if err != nil {
		return err
	}

	for {
		received := make([]byte, 2048)

		n, err := conn.Read(received)

		if n == 0 {
			conn.Close()
			fmt.Println("Disconnected from the server")
			break
		}

		if err != nil {
			fmt.Println("Error receiving message: ", err)
			continue
		}

		msg := received[:n]

		resp := ServerData{}
		json.Unmarshal(msg, &resp)

		ClearTerminal()

		switch resp.Status {
		case "STARTING":
			{
				for i := 0; i < 5; i++ {
					fmt.Printf("Game starting in %d\n", 5-i)
					time.Sleep(1 * time.Second)
				}
			}
		case "ZONES":
			parseZones(resp)
			sel := SelectScreen()
			conn.Write([]byte(fmt.Sprintf("%d", sel)))

		case "SELITEM":
			parseUpgrades(resp)
			log.Println(zoneUpgrades)

		case "ERROR":
			fmt.Println(resp.Message)

		}
	}
	return nil
}

func parseUpgrades(resp ServerData) {
	zoneUpgrades = make([]Upgrade, cap(zoneUpgrades))
	for _, u := range resp.Upgrades {
		zoneUpgrades = append(zoneUpgrades, u)
	}
}

func main() {

	fmt.Print("Choose your username: ")
	in := bufio.NewReader(os.Stdin)

	username, _ := in.ReadString('\n')

	client := NewClient("localhost:3000", username)
	log.Fatal(client.StartClient())
}
