package commute

import (
	"fmt"
	"math/rand"
	"os/exec"
	"time"
)

const DRIVER_STATE = 1
const RIDER_STATE = 2

//One is a complete data structure and other is used just for authentication. While we can use only one
//it may be needed to put TTL and other constraints on auth later.
var gStateDS map[string]*CommState
var gLoggedInUsers map[string]string

//CommState struct basically holds the current set of a commuter. Geo location
//whether she is already connected to a driver/rider etc.
type CommState struct {
	lat         float64
	lng         float64
	curr_state  int
	lastUptTime int64

	//arrReqs is a the pending requests from co-commuters since the last time state was refreshed
	arrReqs []int
	//arrConnectedWith is the list of co-commuters the current user is tied to.
	arrConnectedWith []int
}

//puts a new user into the token data structures and returns the token for auth
func newToken(userName string) string {
	//If user already exists, return as is.
	if val, ok := gLoggedInUsers[userName]; ok {
		return val
	}
	//TODO - do we need to sync here or use channels.

	//Leverage linux command
	newToken := ""
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		newToken = fmt.Sprintf("SomeString:", r.Int63())
	} else {
		newToken = string(out)
	}

	gLoggedInUsers[userName] = newToken
	return newToken

}

func countLoggedInUsers() int {
	return len(gLoggedInUsers)
}

func countStateUsers() int {
	return len(gStateDS)
}

//updateState is invoked every time an event is received from a commuter, be it a heardbeat event or a
//a specific request like "connect ot his driver"
func updateState(userName string, lat float64, lng float64, token string) string {
	currToken := "" //Of the user logged in
	if val, ok := gLoggedInUsers[userName]; ok {
		currToken = string(val)
	}

	if currToken != token {
		fmt.Println("ERROR! Token mismatch. User:", userName, " currToken:", currToken, " token:", token)
		return "Authentication error! you are not logged in"
	}

	var currState *CommState
	if currState2, ok := gStateDS[userName]; ok == false {
		currState = &CommState{}
		currState.arrReqs = make([]int, 0)
		currState.arrConnectedWith = make([]int, 0)
		gStateDS[userName] = currState
	} else {
		currState = currState2
	}

	currState.lastUptTime = time.Now().Unix()
	currState.lat = lat
	currState.lng = lng
	return "Update Success!"

}

//Get the current value as is stored.
func getCurrentState(userName string) *CommState {
	return gStateDS[userName]
}
