package main

import (
	"fmt"
	"log"
	"net/http"
)

const VERSION = "0.1.0"

func main() {
	http.HandleFunc("/", home)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Signal Server %s\n", VERSION)
}
