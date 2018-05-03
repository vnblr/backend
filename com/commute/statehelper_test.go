package commute

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func TestUserCreation(t *testing.T) {
	Initialize()

	//Repeat users
	token1 := newToken("newuser1")
	token2 := newToken("newuser1")

	if token1 != token2 || countLoggedInUsers() != 1 {
		t.Errorf("newuser for same user failed. token1:", token1, " token2:", token2, " size:", countLoggedInUsers())
	}

	token3 := newToken("newuser3")
	if token1 == token3 || countLoggedInUsers() != 2 {
		t.Errorf("newuser for same user failed. token1:", token1, " token3:", token3, " size:", countLoggedInUsers())
	}

}

func TestUpdateAuthError(t *testing.T) {
	Initialize()

	//Not logged in user
	token1 := newToken("newuser1")
	_, err := updateState("token3", 7.1, 10.2, token1, RIDER_STATE, "", EVENT_HEARTBEAT)        //wrong user
	_, err2 := updateState("token1", 7.1, 10.2, "wrongtoken", RIDER_STATE, "", EVENT_HEARTBEAT) //wrong token

	if strings.Contains(err.Error(), "Authentication error") != true ||
		strings.Contains(err2.Error(), "Authentication error") != true {
		t.Errorf("Auth errors were not returned. retStr:", err.Error(), " retStr2:", err.Error())
	}
}

func TestUpdateExample(t *testing.T) {
	Initialize()

	//Lets simulate a login event first.
	token1, _ := updateState("newuser1", 7.1, 10.2, "", RIDER_STATE, "", EVENT_LOGIN)
	if countLoggedInUsers() != 1 || countStateUsers() != 1 { //fixme
		t.Errorf("Count mismatch in DS1. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers())
	}
	r1, _ := updateState("newuser1", 7.1, 10.2, token1, RIDER_STATE, "", EVENT_HEARTBEAT)
	if countLoggedInUsers() != 1 || countStateUsers() != 1 {
		t.Errorf("Count mismatch in DS2. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers(), " ret:", r1)
	}
	//Another update
	r2, _ := updateState("newuser1", 7.2, 10.3, token1, RIDER_STATE, "", EVENT_HEARTBEAT)
	if countLoggedInUsers() != 1 || countStateUsers() != 1 {
		t.Errorf("Count mismatch in DS3. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers(), " ret:", r2)
	}

	//Confirm update
	obj1 := getCurrentState("newuser1")
	if obj1.lat != 7.2 || obj1.lng != 10.3 {
		t.Errorf("Normal update did not work. Users:", countLoggedInUsers(), " lat:", obj1.lat, " lng:", obj1.lng)
	}

	//Nonexistent user
	obj2 := getCurrentState("nonuser")
	if obj2 != nil {
		t.Errorf("nonExistent user get did not work. Users:", countLoggedInUsers(), " lat:", obj2.lat, " lng:", obj2.lng)
	}
}

func createAndUpdateUser(userName string, t *testing.T) {
	defer wg.Done()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	//Simulate a login first..
	token, _ := updateState(userName, 7.1, 10.2, "", RIDER_STATE, "", EVENT_LOGIN)

	//Fire some random updates
	retVal := ""
	for i := 1; i < 1000; i++ {
		retVal, _ = updateState(userName, 1.2, 2.2, token, RIDER_STATE, "", EVENT_HEARTBEAT)
		//Let parallalism kick in
		time.Sleep(1 * time.Millisecond)
	}
	//Final update
	lat := r.Float64()
	lng := r.Float64()
	retVal, _ = updateState(userName, lat, lng, token, RIDER_STATE, "", EVENT_HEARTBEAT)

	time.Sleep(100 * time.Millisecond)
	currState := getCurrentState(userName)
	if currState.lat != lat || currState.lng != lng {
		t.Errorf("createAndUpdateUser error: lat/lng:", lat, lng, " set lat/lng", currState.lat, currState.lng, " r:", retVal)
	}

}

//Lets add many users from across threads
func TestMultithreads(t *testing.T) {
	Initialize()
	numThreads := 1000 //Every go invokation is not a physical thread..but still is a proxy for concurrency.

	for i := 1; i <= numThreads; i++ {
		wg.Add(1)
		go createAndUpdateUser(fmt.Sprintf("MyUser", i), t)
	}
	fmt.Println("TestMultithreads: waiting for threads to come out.", time.Now())
	wg.Wait()
	fmt.Println("TestMultithreads: threads came out.", time.Now())

	if countLoggedInUsers() != numThreads || countStateUsers() != numThreads {
		t.Errorf("Error in TestMultithreads: final counts logged users:", countLoggedInUsers(), " state users:", countStateUsers())
	}

}

//Checks if the nearby drivers/riders logic is working
func TestSearchDrivers(t *testing.T) {
	Initialize()

	cases := []struct {
		user     string
		lat, lng float64
	}{
		{"user1", 100.001, 200.002},
		{"user2", 100.001, 200.003},
		{"user3", 100.001, 200.004},
		{"user4", 100.001, 200.005},
		{"user5", 100.001, 200.002},
		{"user6", 100.001, 200.002},
		{"user7", 150.001, 230.002},
		{"user8", 150.002, 230.001},
	}
	//First update the DS..login these users.
	for _, c := range cases {
		_, _ = updateState(c.user, c.lat, c.lng, "", DRIVER_STATE, "", EVENT_LOGIN)
	}
	//testuser:for now dump in any location
	testToken, _ := updateState("testuser", 10.0, 20.0, "", RIDER_STATE, "", EVENT_LOGIN)

	var retArr []matchUserDetails
	var err error

	//search should give error
	retArr, err = searchMatches("doesntexist", RIDER_STATE)
	if err == nil {
		t.Errorf("No error in searchMatches!")
	}

	//search should result in no nearby users
	_, _ = updateState("testuser", 10.0, 20.0, testToken, RIDER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", RIDER_STATE)
	if len(retArr) != 0 {
		t.Errorf("Error in searchUsers. len retArr:", len(retArr))
	}

	//search should return max possible users
	_, _ = updateState("testuser", 100.001, 200.003, testToken, RIDER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", RIDER_STATE)
	if len(retArr) != MAX_MATCHED_USERS {
		t.Errorf("Error in searchUsers max. len retArr:", len(retArr))
	}
	//search should return 2 possible users
	_, _ = updateState("testuser", 150.001, 230.003, testToken, RIDER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", RIDER_STATE)
	if len(retArr) != 2 {
		t.Errorf("Error in searchUsers max. len retArr:", len(retArr))
	}

}

//Checks if the nearby drivers/riders logic is working
func TestSearchRiders(t *testing.T) {
	Initialize()

	cases := []struct {
		user     string
		lat, lng float64
	}{
		{"user1", 100.001, 200.002},
		{"user2", 100.001, 200.003},
		{"user3", 100.001, 200.004},
		{"user4", 100.001, 200.005},
		{"user5", 100.001, 200.002},
		{"user6", 100.001, 200.002},
		{"user7", 150.001, 230.002},
		{"user8", 150.002, 230.001},
	}
	//First update the DS .. login these users
	for _, c := range cases {
		_, _ = updateState(c.user, c.lat, c.lng, "", RIDER_STATE, "", EVENT_LOGIN)
	}
	//testuser:for now dump in any location
	testToken, _ := updateState("testuser", 10.0, 20.0, "", DRIVER_STATE, "", EVENT_LOGIN)
	//Lets fill it with all requests
	currState := getCurrentState("testuser")
	for _, c := range cases {
		currState.arrReqs = append(currState.arrReqs, c.user)
	}
	var retArr []matchUserDetails

	//search should result in no nearby users
	updateState("testuser", 10.0, 20.0, testToken, DRIVER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", DRIVER_STATE)
	if len(retArr) != 0 {
		t.Errorf("Error in searchUsers. len retArr:", len(retArr))
	}

	//search should return max possible users
	updateState("testuser", 100.001, 200.003, testToken, DRIVER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", DRIVER_STATE)
	if len(retArr) != MAX_MATCHED_USERS {
		t.Errorf("Error in searchUsers max. len retArr:", len(retArr))
	}
	//search should return 2 possible users
	updateState("testuser", 150.001, 230.003, testToken, DRIVER_STATE, "", EVENT_HEARTBEAT)
	retArr, _ = searchMatches("testuser", DRIVER_STATE)
	if len(retArr) != 2 {
		t.Errorf("Error in searchUsers max. len retArr:", len(retArr))
	}

}

//Test a complete sequence once login, search, request, accept. This is not a unit-test per se but a
//overall check
func TestSearchJoinAccept(t *testing.T) {
	//Login a driver
	tokenDriver, _ := updateState("driver1", 100.001, 200.004, "", DRIVER_STATE, "", EVENT_LOGIN)
	//Login a rider
	tokenRider, _ := updateState("rider1", 100.002, 200.001, "", RIDER_STATE, "", EVENT_LOGIN)

	//Lets search for nearby drivers.
	retStr, err := updateState("rider1", 100.002, 200.001, tokenRider, RIDER_STATE, "", EVENT_HEARTBEAT)
	if err != nil || !strings.Contains(retStr, "driver1") {
		t.Errorf("Error in TestSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//Lets send a join request
	retStr, err = updateState("rider1", 100.002, 200.001, tokenRider, RIDER_STATE, "driver1", EVENT_JOINREQ)
	if err != nil || !strings.Contains(retStr, "Success") {
		t.Errorf("Error in TestSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//Let the driver accept it
	retStr, err = updateState("driver1", 100.002, 200.001, tokenDriver, DRIVER_STATE, "rider1", EVENT_JOINACCEPT)
	if err != nil || !strings.Contains(retStr, "Success") {
		t.Errorf("Error in TestSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//Lets confirm the DS settings..
	driverState := getCurrentState("driver1")
	riderState := getCurrentState("rider1")
	if len(driverState.arrReqs) != 0 && len(driverState.arrConnectedWith) != 1 &&
		len(riderState.arrReqs) != 0 && len(riderState.arrConnectedWith) != 1 {
		t.Errorf("Error in TestSearchJoinAccept: ",
			"driver.req:", len(driverState.arrReqs), " driver.conn:", len(driverState.arrConnectedWith),
			"rider.req:", len(riderState.arrReqs), " rider.conn:", len(riderState.arrConnectedWith))

	}

}

//Test a complete sequence once login, search, request, accept. This is not a unit-test per se but a
//overall check
func TestMuiltiSearchJoinAccept(t *testing.T) {
	//Login a driver
	_, _ = updateState("driver1", 100.001, 200.004, "", DRIVER_STATE, "", EVENT_LOGIN)
	//Login a rider
	tokenRider, _ := updateState("rider1", 100.002, 200.001, "", RIDER_STATE, "", EVENT_LOGIN)
	tokenRider2, _ := updateState("rider2", 100.001, 200.008, "", RIDER_STATE, "", EVENT_LOGIN)

	//Lets search for nearby drivers.
	retStr, err := updateState("rider1", 100.002, 200.001, tokenRider, RIDER_STATE, "", EVENT_HEARTBEAT)
	if err != nil || !strings.Contains(retStr, "driver1") {
		t.Errorf("Error in TestMuiltiSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//Lets send a join request
	retStr, err = updateState("rider1", 100.002, 200.001, tokenRider, RIDER_STATE, "driver1", EVENT_JOINREQ)
	if err != nil || !strings.Contains(retStr, "Success") {
		t.Errorf("Error in TestMuiltiSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//The other rider too sees the driver
	retStr, err = updateState("rider2", 100.002, 200.001, tokenRider2, RIDER_STATE, "driver1", EVENT_JOINREQ)
	if err != nil || !strings.Contains(retStr, "Success") {
		t.Errorf("Error in TestMuiltiSearchJoinAccept: err:", err.Error(), " retStr:", retStr)
	}

	//Lets confirm the DS settings..
	driverState := getCurrentState("driver1")
	if len(driverState.arrReqs) != 2 && len(driverState.arrConnectedWith) != 0 {
		t.Errorf("Error in TestMuiltiSearchJoinAccept: ",
			"driver.req:", len(driverState.arrReqs), " driver.conn:", len(driverState.arrConnectedWith))

	}

}
