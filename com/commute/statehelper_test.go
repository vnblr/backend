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
	retStr := updateState("token3", 7.1, 10.2, token1)        //wrong user
	retStr2 := updateState("token1", 7.1, 10.2, "wrongtoken") //wrong token

	if strings.Contains(retStr, "Authentication error") != true ||
		strings.Contains(retStr2, "Authentication error") != true {
		t.Errorf("Auth errors were not returned. retStr:", retStr, " retStr2:", retStr2)
	}
}

func TestUpdateExample(t *testing.T) {
	Initialize()
	token1 := newToken("newuser1")
	if countLoggedInUsers() != 1 || countStateUsers() != 0 { //fixme
		t.Errorf("Count mismatch in DS1. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers())
	}
	fmt.Println("TestUpdateExample here....remove. cnt:", countStateUsers()) //fixme
	r1 := updateState("newuser1", 7.1, 10.2, token1)
	if countLoggedInUsers() != 1 || countStateUsers() != 1 {
		t.Errorf("Count mismatch in DS2. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers(), " ret:", r1)
	}
	//Another update
	r2 := updateState("newuser1", 7.2, 10.3, token1)
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
	token := newToken(userName)

	//Fire some random updates
	retVal := ""
	for i := 1; i < 1000; i++ {
		retVal = updateState(userName, 1.2, 2.2, token)
	}
	//Final update
	lat := r.Float64()
	lng := r.Float64()
	retVal = updateState(userName, lat, lng, token)

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
