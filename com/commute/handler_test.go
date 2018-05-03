package commute

import (
	"testing"
)

func TestProcessReqs(t *testing.T) {
	Initialize()
	cases := []struct {
		username, latlng, mode, etype, retstr, errstr string
	}{
		{"user1", "1.234,-3.4444", "1", "1", "pass", ""},
		{"user1", "wronglat,-3.4444", "1", "1", "", "fail"},
		{"user1", "1.234,wronglng", "1", "1", "", "fail"},
		{"user1", "1.234,-3.4444", "3", "1", "", "wrongmode"},
		{"user1", "1.234,-3.4444", "1", "6", "", "wrongetype"},
	}
	for idx, c := range cases {
		gotstr, err := processRequest(c.username, c.latlng, c.mode, "", "", c.etype)
		if c.errstr == "" && err != nil {
			t.Errorf("test case #", idx, " : processRequest returned error when there was none. ", err.Error())
		}
		if c.errstr != "" && err == nil {
			t.Errorf("test case #", idx, " : processRequest did not returm error when there was one. got:", gotstr)
		}
	}
}
