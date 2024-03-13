package modules

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var zones map[int]Zone
var gameUsers []GameUser

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
	entry, ok := zones[zid]
	if ok {
		entry.users = append(entry.users, *gu)
		zones[zid] = entry
	}
}

type Zone struct {
	users   []GameUser
	upgrade []Upgrade
}

func (z *Zone) RemoveUpgrade(u Upgrade) {
	for i, zi := range z.upgrade {
		if zi == u {
			z.upgrade = append(z.upgrade[:i], z.upgrade[i+1:]...)
		}
	}
}

func Fight(us []GameUser) {
}

func (z *Zone) ProcessZone() {
	log.Println(z.upgrade)
	if len(z.users) > 1 {
		Fight(z.users)
	}

	z.users[0].user.conn.Write(ServerInfo{Status: "SELITEM", Upgrades: z.upgrade}.ToJson())

}

type Upgrade struct {
	Defence int `json:"defence"`
	Attack  int `json:"attack"`
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
		if len(gameUsers) == 1 {
			*endch <- users[0]
		}
	}
}

func generateZones() {
	zones = make(map[int]Zone)

	for i := 0; i < len(gameUsers)*2; i++ {
		z := Zone{}

		for j := 0; j < rand.Intn(5); j++ {
			z.upgrade = append(z.upgrade, Upgrade{Defence: rand.Intn(3), Attack: rand.Intn(4)})
		}

		zones[i] = z
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

	for i := range zones {
		if len(zones[i].users) == 0 {
			continue
		}
		z := zones[i]
		go z.ProcessZone()
	}
}
