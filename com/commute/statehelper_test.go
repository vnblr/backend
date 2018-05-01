package commute

import (
	"strings"
	"testing"
)

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
	if countLoggedInUsers() != 1 && countStateUsers() != 0 {
		t.Errorf("Count mismatch in DS1. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers())
	}
	r1 := updateState("newuser1", 7.1, 10.2, token1)
	if countLoggedInUsers() != 1 && countStateUsers() != 1 {
		t.Errorf("Count mismatch in DS2. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers(), " ret:", r1)
	}
	//Another update
	r2 := updateState("newuser1", 7.2, 10.3, token1)
	if countLoggedInUsers() != 1 && countStateUsers() != 1 {
		t.Errorf("Count mismatch in DS3. LoggedIn:", countLoggedInUsers(), " in DS:", countStateUsers(), " ret:", r2)
	}
}
