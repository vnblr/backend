package commute

import (
	"fmt"
)

type nearbyUserDetails struct {
	userName string
	lat      float64
	lng      float64
	dist     float64 //Already computed, might as well reuse in app
}

//ResponseDetails captures the content of what gets returned by the API.
//For now, its a bunch "already connected commuters" and "nearby commuters". We can iterate on semantics later
//Actually a bit messy: for drivers, it is "requested from" list rather than nearby. Cleanup later.
type ResponseDetails struct {
	//Already connected
	arrConnectedUsers []string

	//Potential connects
	arrNearbyCommuters []nearbyUserDetails
}

func newResponseDetails() *ResponseDetails {

	r := ResponseDetails{}
	r.arrNearbyCommuters = make([]nearbyUserDetails, 0)
	r.arrConnectedUsers = make([]string, 0)
	return &r

}

//Need to convert details to JSON or string. For now, using string
func (r *ResponseDetails) toString(state int) string {
	//Format is driverresppayload,numberofJoinedRiders,rider1,rider2..,numberofrequestedriders,rider1,lat1,lng1,rider2,lat2,lng2..
	//Format is riderresppayload,numberofJoinedDrivers,driver1,driver2..,numberofnearbydrivers,driver1,lat1,lng1,driver2,lat2,lng2..

	retStr := ""
	switch state {
	case DRIVER_STATE:
		retStr = "driverresppayload"
	case RIDER_STATE:
		retStr = "riderresppayload"
	}
	retStr = fmt.Sprintf("%s,%d", retStr, len(r.arrConnectedUsers))
	for _, c := range r.arrConnectedUsers {
		retStr = fmt.Sprintf("%s,%s", retStr, c)
	}
	retStr = fmt.Sprintf("%s,%d", retStr, len(r.arrNearbyCommuters))
	for _, n := range r.arrNearbyCommuters {
		retStr = fmt.Sprintf("%s,%s,%.2f,%.2f", retStr, n.userName, n.lat, n.lng)
	}

	return retStr

}

func (r *ResponseDetails) addJoinedUser(userName string) {
	r.arrConnectedUsers = append(r.arrConnectedUsers, userName)
}

func (r *ResponseDetails) addPotentialUser(userName string, lat float64, lng float64, dist float64) {
	r.arrNearbyCommuters = append(r.arrNearbyCommuters, nearbyUserDetails{userName, lat, lng, dist})
}
