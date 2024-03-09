package modules

import (
	"fmt"
	"log"
	"net"
)

type Client struct {
    address string
    username string
}

func NewClient(address, username string) *Client {

    return &Client{
        address: address,
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
            log.Println("Disconnected from the server")
            break
        }

        if err != nil {
            log.Println("Error receiving message: ", err)
            continue
        }

        fmt.Print("\033[H\033[2J")
        log.Println(string(received[:n]))
    }

    return nil
}
