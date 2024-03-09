package modules

import (
	"fmt"
	"log"
	"net"
	"strings"
)

var users []User 

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
	conn.Write([]byte("Welcome " + username))

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

    for i, u := range inGameUsers {
        if u == user {
            inGameUsers = append(inGameUsers[:i], inGameUsers[i+1:]...)
            break
        }

    }
    log.Println("User", user.username, "disconnected!")
}

type Server struct {
	address string
	port    string
	ln      net.Listener
	qch     chan (struct{})
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

		username := string(usernameBuf[:n])

        user, res := NewUser(conn, username)

        if (res) {
            conn.Write([]byte("Username already taken"))
            conn.Close()
            continue
        }

		go server.readLoop(user)
	}

}

func (server *Server) readLoop(u User) {

    conn := u.conn

	defer conn.Close()

	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if n == 0 {
            u.Disconnect()
			return
		}

		if err != nil {
			log.Println("Error while reading: ", err)
			continue
		}

		msg := string(buf[:n])
        msg = strings.Replace(msg, "\n", "", 1)

		fmt.Println(string(msg))
	}
}
