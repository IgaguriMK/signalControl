package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const VERSION = "0.1.0"

var sections map[string]string

func main() {
	sections = make(map[string]string)

	router := httprouter.New()
	router.GET("/", Home)
	router.GET("/version", Version)
	router.GET("/section/:sectionName", GetSection)
	router.POST("/section/:sectionName", PostSection)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, "<html>")
	fmt.Fprintln(w, "<head>")
	fmt.Fprintf(w, "<title>Signal Server %s</title>\n", VERSION)
	fmt.Fprintln(w, "</head>")
	fmt.Fprintln(w, "<body>")
	fmt.Fprintf(w, "Signal Server %s\n", VERSION)
	fmt.Fprintln(w, "</body>")
}

func Version(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Signal Server %s\n", VERSION)
}

func GetSection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sectionName := ps.ByName("sectionName")
	state, ok := sections[sectionName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404:", sectionName, "Not Found")
		return
	}

	fmt.Fprintln(w, state)
	fmt.Println("Get section:", state)
}

func PostSection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sectionName := ps.ByName("sectionName")
	state := r.PostFormValue("state")
	sections[sectionName] = state
	fmt.Printf("Post section: section[%s] = %s\n", sectionName, state)
}
