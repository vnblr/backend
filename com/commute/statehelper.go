package commute

import (
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

const DRIVER_STATE = 1
const RIDER_STATE = 2
const MAX_WAIT_DISTANCE = 500 //max distance in meters which can be between a driver and commuter.
const MAX_MATCHED_USERS = 5   //Max users that can be shown for a match.

//One is a complete data structure and other is used just for authentication. While we can use only one
//it may be needed to put TTL and other constraints on auth later. Right now, we are using a simple global map.
//For optimization sake, we may have to have multiple DS's indexed by location.
var gStateDS map[string]*CommState
var gLoggedInUsers map[string]string

//locks for the above DS. Maybe we should explore sync.map
var gStateLock = sync.RWMutex{}
var gLoggedInUsersLock = sync.RWMutex{}

//CommState struct basically holds the current set of a commuter. Geo location
//whether she is already connected to a driver/rider etc.
type CommState struct {
	lat           float64
	lng           float64
	curr_state    int
	lastUptTime   int64
	driverOrRider int //Mode of the user.

	//arrReqs is a the pending requests from co-commuters since the last time state was refreshed
	arrReqs []int
	//arrConnectedWith is the list of co-commuters the current user is tied to.
	arrConnectedWith []int
}

//puts a new user into the token data structures and returns the token for auth
func newToken(userName string) string {
	//Sync mess since this can be called from many threads.
	gLoggedInUsersLock.Lock()
	defer gLoggedInUsersLock.Unlock()

	//If user already exists, return as is.
	if val, ok := gLoggedInUsers[userName]; ok {
		return val
	}
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

//Returns the auth token. If the user is not logged in, it will return nil.
func getToken(userName string) string {
	gLoggedInUsersLock.RLock()
	defer gLoggedInUsersLock.RUnlock()

	return gLoggedInUsers[userName]
}

func countLoggedInUsers() int {
	gLoggedInUsersLock.RLock()
	defer gLoggedInUsersLock.RUnlock()

	return len(gLoggedInUsers)
}

func countStateUsers() int {
	gStateLock.RLock()
	defer gStateLock.RUnlock()

	return len(gStateDS)
}

//updateState is invoked every time an event is received from a commuter, be it a heardbeat event or a
//a specific request like "connect ot his driver"
func updateState(userName string, lat float64, lng float64, token string, driverorrider int) string {
	//Sync mess since this can be called from many threads. NOTE: do not introduce recursive locks. As in,
	//do NOT call methods which again do a lock(). It panics. https://stackoverflow.com/questions/14670979/recursive-locking-in-go
	//For eg, do not call newToken or getCurrentState from inside this!
	gLoggedInUsersLock.RLock()
	gStateLock.Lock()
	defer gLoggedInUsersLock.RUnlock()
	defer gStateLock.Unlock()

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
		currState.driverOrRider = driverorrider
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
	gStateLock.RLock()
	defer gStateLock.RUnlock()

	return gStateDS[userName]
}

//The below is the struct that is returned when a search happens.
type matchUserDetails struct {
	userName string
	lat      float64
	lng      float64
	dist     float64
}

//Main function which figures out the nearby commuters. In this POC, we are doing a whole scan. Imagine a
//cluster of machines based on location and users indexed as per a 1x1 grid and we can only relevant maps.
func searchMatches(userName string, mode int) ([]matchUserDetails, error) {
	//readlock
	gStateLock.RLock()
	defer gStateLock.RUnlock()

	if mode != DRIVER_STATE && mode != RIDER_STATE {
		return nil, errors.New(fmt.Sprintf("Invalid mode:", mode))
	}

	var arrMatchedUsers []matchUserDetails = make([]matchUserDetails, 0)
	var currState *CommState = nil
	var ok bool
	if currState, ok = gStateDS[userName]; ok == false {
		return nil, errors.New(fmt.Sprintf("User does not exist in DS:", userName))
	}

	//A rider is typically looking all drivers nearby.
	if mode == RIDER_STATE {
		currPoint := Point{Lat: currState.lat, Lon: currState.lng}
		for u, uState := range gStateDS {
			if uState.driverOrRider == DRIVER_STATE { //can match a rider only to a driver
				//They match only if they are at reasonable distance.
				newPoint := Point{Lat: uState.lat, Lon: uState.lng}
				dist := DistanceBetwnPts(currPoint, newPoint)

				if dist > MAX_WAIT_DISTANCE {
					continue
				}

				//Now this is an eligible user. Lets add.
				newUserDetails := matchUserDetails{u, uState.lat, uState.lng, dist}
				arrMatchedUsers = append(arrMatchedUsers, newUserDetails)

				if len(arrMatchedUsers) >= MAX_MATCHED_USERS {
					break //Come out now. Found enough
				}

			}
		}
		return arrMatchedUsers, nil //Normal return
	}
	//Now the user has to be driver.
	return nil, nil

}
