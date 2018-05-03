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
//These are the various eventTypes honoured
const EVENT_LOGIN = 1      //When you start the app
const EVENT_HEARTBEAT = 2  //Every periodic interval like 30secs based on app settings.
const EVENT_JOINREQ = 3    //When a rider issues a join req looking at drivers
const EVENT_JOINACCEPT = 4 //Driver accepts the pending req and connects.
//What state the current commuter is in
const STATE_LOOKING = 1
const STATE_NOT_LOOKING = 2

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
	arrReqs []string
	//arrConnectedWith is the list of co-commuters the current user is tied to.
	arrConnectedWith []string
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

//Takes care of all authentication/logging in etc. First time a user is created
func newUser(userName string, lat float64, lng float64, driverorrider int) string {
	//Lock down the dbs.
	gStateLock.Lock()
	defer gStateLock.Unlock()

	token := newToken(userName)

	//If state does not exist, create one. No issues here since we already have the user logged in.
	var currState *CommState
	if currState2, ok := gStateDS[userName]; ok == false {
		currState = &CommState{}
		currState.arrReqs = make([]string, 0)
		currState.arrConnectedWith = make([]string, 0)
		currState.driverOrRider = driverorrider
		gStateDS[userName] = currState
	} else {
		currState = currState2
	}

	//Initialize the state.
	currState.lastUptTime = time.Now().Unix()
	currState.lat = lat
	currState.lng = lng
	currState.driverOrRider = driverorrider

	return token

}

//If a wrong token is sent, error out
func isUserValid(userName string, token string) (bool, error) {
	//Read locks
	gLoggedInUsersLock.RLock()
	defer gLoggedInUsersLock.RUnlock()

	currToken := "" //Of the user logged in
	if val, ok := gLoggedInUsers[userName]; ok {
		currToken = string(val)
	}

	if currToken != token {
		fmt.Println("ERROR! Token mismatch. User:", userName, " currToken:", currToken, " token:", token)
		return false, errors.New(fmt.Sprintf("Authentication error! you are not logged in"))
	}

	return true, nil

}

//This is just to update the fields in the global state. It is assume the state is already present,
//if not, just error out.
func updateStateAttrs(userName string, lat float64, lng float64, driverorrider int) error {
	//Write locks
	gStateLock.Lock()
	defer gStateLock.Unlock()
	var currState *CommState
	if currState2, ok := gStateDS[userName]; ok == false {
		fmt.Sprintf("ERROR in updateStateAttrs: user does not exist :", userName, " len:", len(gStateDS))
		return errors.New(fmt.Sprintf("Error while updating profile : ", userName, " does not exist!"))
	} else {
		currState = currState2
	}

	//Initialize the state.
	currState.lastUptTime = time.Now().Unix()
	currState.lat = lat
	currState.lng = lng
	currState.driverOrRider = driverorrider
	return nil //All good.
}

//This is a bit annoying since each of the access needs to be read-locked and we cant do it at a higher level
//Lets see if there is a simpler way out later.
func fillAlreadyJoinedAttr(r *ResponseDetails, userName string) error {
	//Read locks
	gStateLock.RLock()
	defer gStateLock.RUnlock()

	var currState *CommState
	if currState2, ok := gStateDS[userName]; ok == false {
		fmt.Sprintf("ERROR in fillAlreadyJoinedAttr: user does not exist :", userName, " len:", len(gStateDS))
		return errors.New(fmt.Sprintf("Error while updating resp profile : ", userName, " does not exist!"))
	} else {
		currState = currState2
	}

	//Now fill the resp obj
	for _, o := range currState.arrConnectedWith {
		r.addJoinedUser(o)
	}

	return nil

}

//The current user is a rider and wants to register a ride with the "other" user who is a driver
func registerReq(userName string, other string) (string, error) {
	//Write locks
	gStateLock.Lock()
	defer gStateLock.Unlock()

	var currState *CommState
	if currState2, ok := gStateDS[other]; ok == false {
		fmt.Sprintf("ERROR in registerReq: user does not exist :", other, " len:", len(gStateDS))
		return "", errors.New(fmt.Sprintf("Error while registering req : ", other, " does not exist!"))
	} else {
		currState = currState2
	}

	//Now lets register request in this state, if possible.
	if len(currState.arrReqs) >= MAX_MATCHED_USERS {
		return "", errors.New(fmt.Sprintf("Error while registering req : ", other, " is already overloaded!"))
	}

	//See if is already registered
	for _, d := range currState.arrReqs {
		if d == userName {
			return "You are already registerd with this driver. Please wait!", nil
		}
	}
	//Finally...register
	currState.arrReqs = append(currState.arrReqs, userName)
	return fmt.Sprintf("You are now registered with: ", other), nil
}

//Mark the two as "connected". Used in display and analytics subsequently
func joinUsers(rider string, driver string) (string, error) {
	//Write locks
	gStateLock.Lock()
	defer gStateLock.Unlock()

	var riderState *CommState
	if tempState, ok := gStateDS[rider]; ok == false {
		fmt.Sprintf("ERROR in joinUsers: user does not exist :", rider, " len:", len(gStateDS))
		return "", errors.New(fmt.Sprintf("Error while joining user : ", rider, " does not exist!"))
	} else {
		riderState = tempState
	}
	var driverState *CommState
	if tempState2, ok := gStateDS[driver]; ok == false {
		fmt.Sprintf("ERROR in joinUsers: user does not exist :", driver, " len:", len(gStateDS))
		return "", errors.New(fmt.Sprintf("Error while joining user : ", driver, " does not exist!"))
	} else {
		driverState = tempState2
	}

	//Now that we have both states, lets update them.
	riderState.arrConnectedWith = append(riderState.arrConnectedWith, driver)
	driverState.arrConnectedWith = append(driverState.arrConnectedWith, rider)

	//Remove request registered. Is there a better way in golang?
	//https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-array-in-golang
	for idx, req := range driverState.arrReqs {
		if req == rider {
			oldArr := driverState.arrReqs
			oldArr[idx] = oldArr[len(oldArr)-1]
			driverState.arrReqs = oldArr[:len(oldArr)-1]
			break
		}
	}
	return "Success in Join operation!", nil

}

//updateState is invoked every time an event is received from a commuter, be it a heardbeat event or a
//a specific request like "connect ot his driver". This is the main router and calls internal methods
//to process request.
func updateState(userName string, lat float64, lng float64, token string, driverorrider int,
	other string, eventType int) (string, error) {
	//A general note. Do not take a mutex at this high a level. Better to do it at granular functions which are
	//eventually routed to. Else will run into issues as in https://stackoverflow.com/questions/14670979/recursive-locking-in-go

	//Ensure eventtype sanity
	if eventType != EVENT_HEARTBEAT && eventType != EVENT_JOINREQ &&
		eventType != EVENT_JOINACCEPT && eventType != EVENT_LOGIN {
		return "", errors.New(fmt.Sprintf("Invalid eventtype", eventType))
	}
	var err error
	var retStr string

	//Step #1: if this is a login event, create a new user
	if eventType == EVENT_LOGIN {
		currToken := newUser(userName, lat, lng, driverorrider)
		//Nothing else to do for now. Just return the newly created token which will be passed back.
		return currToken, nil
	}

	_, err = isUserValid(userName, token)
	if err != nil {
		return "", err
	}

	//Now lets handle the events.

	//Whatever be the event, lets update the location etc first.
	err = updateStateAttrs(userName, lat, lng, driverorrider)
	if err != nil {
		return "", err
	}

	switch eventType {
	case EVENT_HEARTBEAT: //This comes at prefined periodicity from app-side. Maybe once in 30 secs if user is moving
		//Lets find out the nearby commuters and return back.
		var arrMatchUsers []matchUserDetails
		arrMatchUsers, err = searchMatches(userName, driverorrider)
		if err != nil {
			return "", err
		}
		//Instantiate a response details object
		var respObj *ResponseDetails = newResponseDetails()
		err = fillAlreadyJoinedAttr(respObj, userName)
		if err != nil {
			return "", err
		}
		//Now fill the details of matched users
		for _, m := range arrMatchUsers {
			respObj.addPotentialUser(m.userName, m.lat, m.lng, m.dist)
		}
		//Return the response
		return respObj.toString(driverorrider), nil

	case EVENT_JOINREQ: //This comes when a rider specifically asks to join a driver which is displayed on the app.
		retStr, err = registerReq(userName, other)
		if err != nil {
			return "", err
		}
		return retStr, nil //ALl good. Request is registered with the "other" driver.

	case EVENT_JOINACCEPT: //This comes when a driver accepts a request from a nearby rider.
		retStr, err = joinUsers(other, userName) //Note that other=rider in this signal
		if err != nil {
			return "", err
		}
		return retStr, nil //ALl good. Request is registered with the "other" driver.

	}

	return "Update Success!", nil

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

	var arrMatchedUsers []matchUserDetails = make([]matchUserDetails, 0)
	var currState *CommState = nil
	var ok bool
	if currState, ok = gStateDS[userName]; ok == false {
		return nil, errors.New(fmt.Sprintf("User does not exist in DS:", userName))
	}
	if mode != currState.driverOrRider {
		return nil, errors.New(fmt.Sprintf("Invalid mode:", mode))
	}

	currPoint := Point{Lat: currState.lat, Lon: currState.lng}

	//A rider is typically looking all drivers nearby.
	if mode == RIDER_STATE {
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
	//Now the user has to be driver. Here, you just go by riders' requests. Scan through, update latest
	//distance and just return
	for _, reqUser := range currState.arrReqs {

		//For now, if the requested user is not found, we just move on. Ideally we should error out and handle.
		if reqUserState, ok := gStateDS[reqUser]; ok {
			if reqUserState.driverOrRider == RIDER_STATE { //Again, lets ignore if the state is wrong
				newPoint := Point{Lat: reqUserState.lat, Lon: reqUserState.lng}
				dist := DistanceBetwnPts(currPoint, newPoint)

				if dist > MAX_WAIT_DISTANCE {
					continue
				}

				//Now this is an eligible user. Lets add.
				newUserDetails := matchUserDetails{reqUser, reqUserState.lat, reqUserState.lng, dist}
				arrMatchedUsers = append(arrMatchedUsers, newUserDetails)

				if len(arrMatchedUsers) >= MAX_MATCHED_USERS {
					break //Come out now. Found enough
				}

			}
		}
	}

	return arrMatchedUsers, nil

}
