package commute

import (
//"container/list"
)

const DRIVER_STATE = 1
const RIDER_STATE = 2

//commState struct basically holds the current set of a commuter. Geo location
//whether she is already connected to a driver/rider etc.
type commState struct {
	lat         float64
	lng         float64
	curr_state  int
	lastUptTime int64

	//listReqs list
	//listJoinedWith list
}
