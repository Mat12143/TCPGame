package modules

import (
	"fmt"
	"log"
	"time"
)

var inGameUsers []User

func broadcast(msg string) {

	for _, u := range users {

		_, err := u.conn.Write([]byte(msg))
		if err != nil {
			log.Fatal(err)
			continue
		}
	}

}

func checkForEnd(endch *chan (User)) {
	for {
		if len(inGameUsers) == 1 {
            *endch <- inGameUsers[0]
		}
	}
}

func StartLoop() {

	for {
		if len(users) < 3 {
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}

    for i := range 10 {
        str := fmt.Sprintf("Game starting in %d seconds", i + 1)
        broadcast(str)
        time.Sleep(time.Duration(1) * time.Second)

    }

    for _, u := range users {
        inGameUsers = append(inGameUsers, u)
    }

	broadcast("Game started")
    log.Println("Game started")

    winner := make(chan(User))

    go checkForEnd(&winner)

    u := <- winner

    inGameUsers = make([]User, cap(users))

    broadcast("Game ended. Won: " + u.username)
    log.Println("Game ended. Won: " + u.username)

}
