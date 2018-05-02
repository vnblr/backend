package commute

import (
	"testing"
)

func TestProcessReqs(t *testing.T) {
	Initialize()
	cases := []struct {
		username, lat, lng, retstr, errstr string
	}{
		{"user1", "1.234", "-3.4444", "pass", ""},
		{"user1", "wronglat", "-3.4444", "", "fail"},
		{"user1", "1.234", "wronglng", "", "fail"},
	}
	for _, c := range cases {
		gotstr, err := processRequest(c.username, c.lat, c.lng)
		if c.errstr == "" && err != nil {
			t.Errorf("processRequest returned error when there was none. ", err.Error())
		}
		if c.errstr != "" && err == nil {
			t.Errorf("processRequest did not returm error when there was one. lat", c.lat, " lng:", c.lng, " ret:", gotstr)
		}
	}
}
