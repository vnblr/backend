# backend

To build : 
run "go install" from inside cmd folder. It will create binary in $GOPATH/bin/ folder. 

To run unit tests :
run "go test github.com/vnblr/backend/com/commute". To look at coverage do a "go test -coverprofile=/tmp/cover.out" and then "go tool cover -html=/tmp/cover.out "

godoc :
run "godoc -http:6060" and in browser hit http://127.0.0.1:6060/pkg/github.com/vnblr/backend/com/commute/

"go get github.com/vnblr/backend" should clone and install. <TODO>
