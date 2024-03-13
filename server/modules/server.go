package modules

import (
	"encoding/json"
	"log"
	"net"
	"strings"
)

var users []User

type ServerInfo struct {
	Status   string    `json:"status"`
	Zones    int       `json:"zones"`
	Winner   string    `json:"winner"`
	Message  string    `json:"message"`
	Upgrades []Upgrade `json:"upgrades"`
}

func (si ServerInfo) ToJson() []byte {

	r, err := json.Marshal(si)
	if err != nil {
		return []byte{}
	}
	return r
}

type User struct {
	conn     net.Conn
	username string
}

func NewUser(conn net.Conn, username string) (User, bool) {

	for _, u := range users {
		if u.username == username {
			return User{}, true
		}
	}

	u := User{conn, username}

	users = append(users, u)

	return u, false
}

func (user User) Disconnect() {
	defer user.conn.Close()

	for i, u := range users {
		if u == user {
			users = append(users[:i], users[i+1:]...)
			break
		}
	}
	log.Println("User", user.username, "disconnected!")
}

type Server struct {
	address string
	port    string
	ln      net.Listener
}

func NewServer(address, port string) *Server {
	return &Server{
		address: address,
		port:    port,
	}
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

func (server *Server) acceptLoop() {
	for {

		conn, err := server.ln.Accept()
		if err != nil {
			log.Println("Accept error: ", err)
			continue
		}

		log.Println("New connection from ", conn.RemoteAddr())

		usernameBuf := make([]byte, 2048)

		n, err := conn.Read(usernameBuf)

		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}

		username := string(usernameBuf[:n])
		username = strings.ReplaceAll(username, "\n", "")

		_, res := NewUser(conn, username)

		if res {
			conn.Write(ServerInfo{Status: "ERROR", Message: "Username already taken"}.ToJson())
			conn.Close()
			continue
		}
	}
}
