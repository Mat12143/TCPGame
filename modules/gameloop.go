package modules

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

var inGameUsers []User
var inGameZones map[string]string
var zones []Zone
var items = []Item{Item{id: "sword", stats: ItemStats{defence: 0, attack: 1}}, Item{id: "armour", stats: ItemStats{defence: 1, attack: 0}}}

type Zone struct {
	name  string
	users []User
	items []Item
}

type Item struct {
	id       string
	stats    ItemStats
	quantity int
}

type ItemStats struct {
	defence int
	attack  int
}

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

func GenerateZones() {

	for range 3 {
		for range 2 {

			var zItems []Item

			for range rand.Intn(4) {
				i := items[rand.Intn(len(items)-1)]
				i.quantity = rand.Intn(3)
				zItems = append(zItems, i)
			}

			z := Zone{
				name:  "Testing land",
				items: zItems,
			}

			zones = append(zones, z)
		}

	}
}

func receiveZones(u User, ich *chan(string)) {

    conn := u.conn

    buf := make([]byte, 2048)
    n, _ := conn.Read(buf)

    msg := string(buf[:n])
    msg = strings.Replace(msg, "\n", "", 1)

    if strings.Contains(msg, "SEL ") {
        *ich <- u.username + "|" + strings.Replace(msg, "SEL ", "", 1)
        
    }
}

func turn() {

    zoneStr := ""
    for _, z := range zones {
        zoneStr += z.name + " "
    }

    broadcast("CHOOSE " + zoneStr)

    ich := make(chan(string))

    for _, u := range inGameUsers {
        go receiveZones(u, &ich)
    }

    go func(ich *chan(string)){
        time.Sleep(5)
        close(*ich)
    }(&ich)

    for {
        sel := <-ich
        inGameZones[strings.Split(sel, "|")[0]] = strings.Split(sel, "|")[1]
    }
}

func StartLoop() {

	for {
		if len(users) >= 1 {
			break
		} else {
			time.Sleep(time.Second * 2)
		}
	}

	for i := range 10 {
		broadcast(fmt.Sprintf("Game starting in %d seconds", 10-i))
		time.Sleep(time.Duration(1) * time.Second)
	}

	for _, u := range users {
		inGameUsers = append(inGameUsers, u)
	}

	broadcast("START")
	log.Println("Game started")

    GenerateZones()

    turn()

    log.Println(inGameZones)

	winner := make(chan (User))

	u := <-winner

	inGameUsers = make([]User, cap(users))

    broadcast("END " + u.username)
	log.Println("Game ended. Won: " + u.username)
}
