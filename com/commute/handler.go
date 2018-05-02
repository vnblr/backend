package commute

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

var reqCh chan int
var reqCnt int = 0

//For the sake of a unit test
func dummyFun(str string) string {
	var retStr string
	retStr = "SomeStr:" + str
	return retStr
}

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

	fmt.Println(time.Now(), "\t", user, "\t", ip, "\t", lat, "\t", lng, "\t", ua,
		"\t", r.URL.RawQuery)

	fmt.Fprintf(w, "Request received :")

}
