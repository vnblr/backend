package commute

import (
	"testing"
)

//Format of the output string
func TestRespOutput(t *testing.T) {
	cases := []struct {
		joinedusers       []string
		mode              int
		users             []string
		lats, lngs, dists []float64
		finalStr          string
	}{
		{joinedusers: []string{"juser1", "juser2"}, users: []string{"cuser1", "cuser2"}, mode: 1,
			lats: []float64{1.1, 2.2}, lngs: []float64{3.1, 4.2}, dists: []float64{100.1, 120.1},
			finalStr: "driverresppayload,2,juser1,juser2,2,cuser1,1.10,3.10,cuser2,2.20,4.20"},
		{joinedusers: []string{"x"}, users: []string{"driver1", "driver2"}, mode: 2,
			lats: []float64{1.1, 2.2}, lngs: []float64{3.1, 4.2}, dists: []float64{100.1, 120.1},
			finalStr: "riderresppayload,1,x,2,driver1,1.10,3.10,driver2,2.20,4.20"},
	}

	for idx, c := range cases {
		obj := newResponseDetails()
		for _, u := range c.joinedusers {
			obj.addJoinedUser(u)
		}
		for idx2, u := range c.users {
			obj.addPotentialUser(u, c.lats[idx2], c.lngs[idx2], c.dists[idx2])
		}

		if obj.toString(c.mode) != c.finalStr {
			t.Errorf("Test case #:%d Error in returned string. Expected:%s returned:%s", idx,
				c.finalStr, obj.toString(c.mode))
		}
	}

}
