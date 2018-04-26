package main

import (
	"fmt"
	"github.com/vnblr/backend/com/commute"
	"net/http"
)

//package cmd/main is the entry point to run as a http container.
func main() {
	fmt.Println("MapsBackend : entry point start.")

	commute.Initialize()

	http.HandleFunc("/", commute.Handler)
	http.ListenAndServe(":8080", nil)

	fmt.Println("MapsBackend : Done launching server at 8080")
}
