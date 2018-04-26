package commute

import (
	"fmt"
	"net/http"
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
	//A parallel thread to dump stats
	go printStat()

}

//Function Handler is the entry-point which is registered in the http handler.
//Every http request lands here.
func Handler(w http.ResponseWriter, r *http.Request) {
	reqCh <- 1

	fmt.Fprintf(w, "Request received :")

}
