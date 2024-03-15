package modules

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var zones map[int]Zone
var gameUsers []GameUser
var wg sync.WaitGroup

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
	for _, u := range us {
		u.user.conn.Write([]byte("Combat to be implemented"))
	}
}

func ProcessZone(i int) {
	z := zones[i]
	log.Println(zones)

	defer wg.Done()
	defer func() {
		z.users = []GameUser{}
		zones[i] = z
	}()

	if len(z.users) > 1 {
		Fight(z.users)
		return
	}
	z.users[0].user.conn.Write(ServerInfo{Status: "SELITEM", Upgrades: z.upgrade}.ToJson())

	if len(z.upgrade) == 0 {
		return
	}

	upBuf := make([]byte, 2048)

	n, err := z.users[0].user.conn.Read(upBuf)

	if err != nil {
		log.Fatal(err)
	}

	upID := string(upBuf[:n])

	ID, _ := strconv.Atoi(upID)

	z.users[0].AddUpgrade(z.upgrade[ID])
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

	for {
		log.Println("NEW TURN")
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
			wg.Add(1)
			go ProcessZone(i)
		}
		wg.Wait()
	}
}
