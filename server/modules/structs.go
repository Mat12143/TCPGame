package modules

import (
	"encoding/json"
	"log"
	"net"
)

// Server Struct
type Server struct {
	address string
	port    string
	ln      net.Listener
}

func (server *Server) Start() error {
	ln, err := net.Listen("tcp", server.address+":"+server.port)
	defer ln.Close()

	if err != nil {
		return err
	}

	log.Printf("Server listening on %s:%s\n", server.address, server.port)

	server.ln = ln

	go StartLoop()

	server.acceptLoop()

	return nil
}

// JSON Communications between Server and Client
type ServerCommunication struct {
	Status   string    `json:"status"`
	Zones    int       `json:"zones"`
	Winner   string    `json:"winner"`
	Message  string    `json:"message"`
	Upgrades []Upgrade `json:"upgrades"`
}

func (sc ServerCommunication) ToJson() []byte {
	r, err := json.Marshal(sc)

	if err != nil {
		return []byte{} 
	}

	return r 
}

// Connected user
type User struct {
	conn     net.Conn
	username string
}

func (user User) Disconnect() {
	defer user.conn.Close()

	for i, u := range users {
		if u == user {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}
    for i, u := range gameUsers {
        if u.user == user {
            gameUsers = append(gameUsers[:i], gameUsers[i+1:]...)
            break
        }
    }

	log.Println("User", user.username, "disconnected!")
}

// In-Game user
type GameUser struct {
	user     User
	upgrades []Upgrade
	lifes    int
}

func (gu *GameUser) AddUpgrade(u Upgrade) {
	gu.upgrades = append(gu.upgrades, u)
}

func (gu *GameUser) Die() {

	for i, u := range gameUsers {
		if &u == gu {
			gameUsers = append(gameUsers[:i], gameUsers[i+1:]...)
			break
		}
	}
}

func (gu *GameUser) RemoveLife(points int) {
	gu.lifes = gu.lifes - points
	if gu.lifes <= 0 {
		gu.Die()
	}
}

func (gu *GameUser) AddToZone(zid int) {
	entry, ok := inGameZones[zid]
	if ok {
		entry.users = append(entry.users, gu)
		inGameZones[zid] = entry
	}
}

// Game Zones
type Zone struct {
	users   []*GameUser
	upgrade []Upgrade
}

func (z *Zone) RemoveUpgrade(u Upgrade) {
	for i, zi := range z.upgrade {
		if zi == u {
			z.upgrade = append(z.upgrade[:i], z.upgrade[i+1:]...)
		}
	}
}

// Upgrade
type Upgrade struct {
	Defence int `json:"defence"`
	Attack  int `json:"attack"`
}

type Selection struct {
	user      GameUser
	selection int
}
