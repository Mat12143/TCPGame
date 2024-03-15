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

type Stats struct {
	Attack  int
	Defence int
}

func (s *Stats) add(u Upgrade) {
	*&s.Attack += u.Attack
	*&s.Defence += u.Defence
}

var zones map[int]string
var zoneStrings map[int]string
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

	stats := Stats{Attack: 0, Defence: 0}

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
			sel := SelectScreen(zones)
			conn.Write([]byte(fmt.Sprintf("%d", sel)))

		case "SELITEM":
			parseUpgrades(resp)
			if len(zoneStrings) == 0 {
				fmt.Println("No upgrades in this zone")
				break
			}

			up := SelectScreen(zoneStrings)
			conn.Write([]byte(fmt.Sprintf("%d", up)))

			stats.add(zoneUpgrades[up])

		case "ERROR":
			fmt.Println(resp.Message)
		}

		fmt.Printf("Attack: %d | Defence %d", stats.Attack, stats.Defence)
	}
	return nil
}

func parseUpgrades(resp ServerData) {
	zoneStrings = make(map[int]string)
	zoneUpgrades = make([]Upgrade, cap(zoneUpgrades))

	for i, u := range resp.Upgrades {
		zoneStrings[i] = fmt.Sprintf("+%d Attack | +%d Defence", u.Attack, u.Defence)
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
