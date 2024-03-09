package modules

import (
	"fmt"
	"net"
	"strings"

	"github.com/Mat12143/TCPGame/ui"
)

type Client struct {
	address  string
	username string
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

		msg := string(received[:n])
		msg = strings.Replace(msg, "\n", "", 1)

		if strings.Contains(msg, "CHOOSE") {
			zs := make([]string, cap(zones))

			msgZones := strings.Replace(msg, "CHOOSE ", "", 1)
			for _, z := range strings.Split(msgZones, " ") {
				if len(z) > 0 {

					zs = append(zs, z)
				}
			}

			zone := ui.ZoneSelect(zs)
			conn.Write([]byte("SEL " + zone))
		} else {
			fmt.Println(msg)
		}
	}
	return nil
}
