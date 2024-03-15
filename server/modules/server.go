package modules

import (
	"log"
	"net"
	"strings"
)

var users []User

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

func NewServer(address, port string) *Server {
	return &Server{
		address: address,
		port:    port,
	}
}

func (server *Server) acceptLoop() {
	for {
		conn, err := server.ln.Accept()

		if err != nil {
			log.Println("Accept error: ", err)
			continue
		}

		log.Println("New connection from ", conn.RemoteAddr())

		usernameBuf := make([]byte, MAX_SIZE)

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
			resp := ServerCommunication{Status: "ERROR", Message: "Username already taken"}.ToJson()
			conn.Write(resp)

			conn.Close()
			continue
		}
        log.Printf("New user entered the chat: %s\n", username)
	}
}

func isConnDead(n int, err error) bool {

    if err != nil {
        return true
    }

    if n <= 0 {
        return true
    }

    return false
}
