package commute

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var reqCh chan int
var reqCnt int = 0

//Function printStat is a global Stat counter. It will print hygiene stats like #requests,
// #errs, latency(?) etc. Start with count first
func printStat() {
	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case ci := <-reqCh:
			reqCnt += ci
			if reqCnt%10 == 0 {
				fmt.Println(" ReqCnt : ", reqCnt, " Time taken=", time.Since(start))

			}
		case <-ticker.C:
			fmt.Println(" ReqCnt : ", reqCnt, " Time taken=", time.Since(start))

		}
	}
}

//Function Initialize does all the channel etc initialization and launches threads to monitor
func Initialize() {
	reqCh = make(chan int, 100)
	gStateDS = make(map[string]*CommState, 1000)
	gLoggedInUsers = make(map[string]string, 1000)
	//A parallel thread to dump stats
	go printStat()

}

//entry point calls this...directly accepts strings as given in URL and then does
//parsing and validation. Keep it as is..easier to unit-test.
func processRequest(userName string, latlngstr string,
	driverorrider string, token string, other string,
	eventtype string) (string, error) {
	//Now lets process the params
	var latLongArr []string = strings.Split(latlngstr, ",")
	if len(latLongArr) != 2 {
		return "", errors.New(fmt.Sprintf("latlongstr wrong format:%s", latlngstr))
	}
	var latFlt, lngFlt float64
	var err1, err2, err error
	latFlt, err1 = strconv.ParseFloat(latLongArr[0], 64)
	if err1 != nil {
		return "", errors.New(fmt.Sprintf("ERROR in lat parameter:%s", err1.Error()))
	}
	lngFlt, err2 = strconv.ParseFloat(latLongArr[1], 64)
	if err2 != nil {
		return "", errors.New(fmt.Sprintf("ERROR in lng parameter:%s", err2.Error()))
	}
	var driverRiderMode int
	driverRiderMode, err = strconv.Atoi(driverorrider)
	if err != nil || (driverRiderMode != RIDER_STATE && driverRiderMode != DRIVER_STATE) {
		return "", errors.New(fmt.Sprintf("ERROR in mode parameter:%s", driverorrider))
	}
	var everyTypeParsed int
	everyTypeParsed, err = strconv.Atoi(eventtype)
	if err != nil || (everyTypeParsed != EVENT_LOGIN && everyTypeParsed != EVENT_HEARTBEAT &&
		everyTypeParsed != EVENT_JOINREQ && everyTypeParsed != EVENT_JOINACCEPT) {
		return "", errors.New(fmt.Sprintf("ERROR in eventtype parameter:%s", eventtype))
	}

	//Now hand the thing over to the updater
	retValue, err := updateState(userName, latFlt, lngFlt, token, driverRiderMode, other, everyTypeParsed)
	return retValue, err //Return as is
}

//Function Handler is the entry-point which is registered in the http handler.
//Every http request lands here.
func Handler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, "commute/map") {
		return
	}

	reqCh <- 1

	//Parse the parameters in the request
	ip := r.RemoteAddr
	user := r.URL.Query().Get("user")
	ua := r.Header.Get("User-Agent")
	token := r.URL.Query().Get("token")

	//The following are legacy reasons we are keep params as is. Lets chagne both app and these strings soon.
	latlngstr := r.URL.Query().Get("param") //looks like "100.112,300.117"
	status := r.URL.Query().Get("status")   //This actually is the "other"
	eventtype := r.URL.Query().Get("eventtype")
	driverorrider := r.URL.Query().Get("mode")

	retValue, err := processRequest(user, latlngstr, driverorrider, token, status, eventtype)
	if err != nil {
		fmt.Fprintf(w, "ERROR! :", err)
	} else {
		fmt.Fprintf(w, retValue)
	}

	//fmt.Fprintf(w, "Request processed successfully :")

	//Ideally should be using some logging system. TODO.
	fmt.Println(time.Now(), "\t", user, "\t", ip, "\t", latlngstr, "\t", ua,
		"\t", r.URL.RawQuery, "\t", retValue, "\t", err)

}
