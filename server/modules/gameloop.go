package modules

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var zones map[int]Zone
var gameUsers []GameUser

type GameUser struct {
	user   User
	spells []Spell
	lifes  int
}

func (gu *GameUser) AddSpell(s Spell) {
	gu.spells = append(gu.spells, s)
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
    for i, z := range zones {
        if i == zid {
            z.users = append(z.users, *gu)
        }
    }
}

type Zone struct {
	users  []GameUser
	spells []Spell
}

func (z *Zone) RemoveItem(item Spell) {
	for i, zi := range z.spells {
		if zi == item {
			z.spells = append(z.spells[:i], z.spells[i+1:]...)
		}
	}
}

type Spell struct {
	defence int
	attack  int
}

type Selection struct {
	user      GameUser
	selection int
}

func broadcast(bt []byte) {

	for _, u := range gameUsers {

		_, err := u.user.conn.Write(bt)
		if err != nil {
			log.Fatal(err)
			continue
		}
	}
}

func checkForEnd(endch *chan (User)) {
	for {
		if len(users) == 1 {
			*endch <- users[0]
		}
	}
}

func generateZones() {
	zones = make(map[int]Zone)

	o := 0
	for i := 0; i < len(gameUsers); i++ {
		for y := 0; y < 2; y++ {
			z := Zone{}

			for j := 0; j < rand.Intn(5); j++ {
				z.spells = append(z.spells, Spell{defence: rand.Intn(3), attack: rand.Intn(4)})
			}

			zones[y+i+o] = z
			o += 1
		}
	}
}

func waitForSel(gu GameUser, ch *chan (Selection)) {
	selBuf := make([]byte, 2048)

	n, err := gu.user.conn.Read(selBuf)
	if err != nil {
		log.Fatal(err)
		return
	}
    sel := string(selBuf[:n])
    sel = strings.ReplaceAll(sel, "\n", "")

	i, err := strconv.Atoi(sel)

    if err != nil {
        log.Fatal(err)
        return
    }

	*ch <- Selection{user: gu, selection: i}
}

func StartLoop() {

	for {
		if len(users) >= 2 {
			break
		}
		time.Sleep(time.Second * 2)
	}

	for _, u := range users {
		gu := GameUser{
			user:  u,
			lifes: 2,
		}
		gameUsers = append(gameUsers, gu)
	}

	broadcast(ServerInfo{Status: "STARTING"}.ToJson())

	// Wait for 5 seconds before starting
	time.Sleep(5 * time.Second)
	generateZones()

	broadcast(ServerInfo{Status: "ZONES", Zones: len(zones)}.ToJson())

	selchan := make(chan (Selection))

	for _, u := range gameUsers {
		go waitForSel(u, &selchan)
	}

    for i := 0; i < len(gameUsers); i++ {
        sel := <-selchan
		log.Printf("User %s selected zone %d\n", sel.user.user.username, sel.selection)
        sel.user.AddToZone(sel.selection)
	}

    log.Println("FINISHED")
    fmt.Printf("zones: %v\n", zones)


}
