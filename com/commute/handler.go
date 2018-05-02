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

//entry point calls this.
func processRequest(userName string, lat string, lng string) (string, error) {
	var myToken string
	myToken = getToken(userName)

	if myToken == "" {
		myToken = newToken(userName)
	}

	//Now lets process the params
	var latFlt, lngFlt float64
	var err1, err2 error
	latFlt, err1 = strconv.ParseFloat(lat, 64)
	if err1 != nil {
		return "", errors.New(fmt.Sprintf("ERROR in lat parameter:", err1.Error()))
	}

	lngFlt, err2 = strconv.ParseFloat(lng, 64)
	if err2 != nil {
		return "", errors.New(fmt.Sprintf("ERROR in lng parameter:", err2.Error()))
	}

	//Now hand the thing over to the updater
	retValue := updateState(userName, latFlt, lngFlt, myToken, RIDER_STATE)
	return retValue, nil
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
	lat := r.URL.Query().Get("lat")
	lng := r.URL.Query().Get("lng")
	user := r.URL.Query().Get("user")
	ua := r.Header.Get("User-Agent")

	retValue, err := processRequest(user, lat, lng)
	if err != nil {
		fmt.Fprintf(w, "ERROR! :", err)
	} else {
		fmt.Fprintf(w, retValue)
	}

	fmt.Fprintf(w, "Request received :")
	fmt.Println(time.Now(), "\t", user, "\t", ip, "\t", lat, "\t", lng, "\t", ua,
		"\t", r.URL.RawQuery, "\t", retValue, "\t", err)

}
