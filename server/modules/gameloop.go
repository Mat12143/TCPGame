package modules

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

var inGameZones map[int]Zone
var gameUsers []GameUser
var wg sync.WaitGroup

func getUserByUsername(username string) (GameUser, int) {
	for i, us := range gameUsers {
		if us.user.username == username {
			return us, i
		}
	}
	return GameUser{}, 0
}

func shuffleSlice(us []string) []string {
	for i := range us {
		j := rand.Intn(i + 1)
		us[i], us[j] = us[j], us[i]
	}
	return us
}

func Fight(usernames []string) {
    for _, us := range usernames {
        user, _ := getUserByUsername(us)
        n, err := user.user.conn.Write(ServerCommunication{Status: "FIGHT", Message: "Fight to be implemented"}.ToJson())
        if isConnDead(n, err) {
            user.user.Disconnect()
        }
    }
}

func ProcessZone(i int) {
	z := inGameZones[i]

	defer wg.Done()
	defer func() {
		// Reset zone users slice on zone processing end
		z.users = []string{}
		inGameZones[i] = z
	}()

	if len(z.users) > 1 {
		Fight(z.users)
		return
	}

	user, id := getUserByUsername(z.users[0])

	n, err := user.user.conn.Write(ServerCommunication{Status: "SELITEM", Upgrades: z.upgrade}.ToJson())

	if isConnDead(n, err) {
		user.user.Disconnect()
		return
	}

	// If there are no upgrades not need to wait
	if len(z.upgrade) == 0 {
		return
	}

	respBuf := make([]byte, 2048)

	n, err = user.user.conn.Read(respBuf)

	if isConnDead(n, err) {
		user.user.Disconnect()
		return
	}

	upID := string(respBuf[:n])
	ID, _ := strconv.Atoi(upID)

	user.AddUpgrade(z.upgrade[ID])
	gameUsers[id] = user
}

func broadcast(bt []byte) {

	for _, u := range gameUsers {

		n, err := u.user.conn.Write(bt)
		if isConnDead(n, err) {
			u.user.Disconnect()
			continue
		}
	}
}

func checkForEnd() GameUser {
	if len(gameUsers) <= 1 {
		return gameUsers[0]
	}
	return GameUser{}
}

func generateZones() {
	inGameZones = make(map[int]Zone)

	for i := 0; i < len(gameUsers)*2; i++ {
		z := Zone{}

		for j := 0; j < rand.Intn(5); j++ {
			z.upgrade = append(z.upgrade, Upgrade{Defence: rand.Intn(3) + 1, Attack: rand.Intn(4) + 1})
		}

		inGameZones[i] = z
	}
}

func waitForSel(gu GameUser, ch *chan (Selection)) {
	selBuf := make([]byte, 2048)

	n, err := gu.user.conn.Read(selBuf)

	if isConnDead(n, err) {
		gu.user.Disconnect()
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
		if len(users) >= PLAYER_START {
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

	broadcast(ServerCommunication{Status: "STARTING"}.ToJson())

	// Wait for 5 seconds before starting
	time.Sleep(5 * time.Second)
	generateZones()

	for {
		wu := checkForEnd()
		if wu.lifes != 0 {
			broadcast(ServerCommunication{Status: "WINNER", Winner: wu.user.username}.ToJson())
			break
		}

		broadcast(ServerCommunication{Status: "ZONES", Zones: len(inGameZones)}.ToJson())

		selchan := make(chan (Selection))

		for _, u := range gameUsers {
			go waitForSel(u, &selchan)
		}

		for i := 0; i < len(gameUsers); i++ {
			sel := <-selchan
			log.Printf("User %s selected zone %d\n", sel.user.user.username, sel.selection)
			sel.user.AddToZone(sel.selection)
		}

		for i := range inGameZones {
			if len(inGameZones[i].users) == 0 {
				continue
			}
			wg.Add(1)
			go ProcessZone(i)
		}
		wg.Wait()
		log.Println(gameUsers)

        broadcast(ServerCommunication{Status: "TURN"}.ToJson())
        time.Sleep(5 * time.Second)
	}
}
