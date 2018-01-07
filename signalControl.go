package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

const VERSION = "0.1.0"

var db *sql.DB

func main() {
	LodaConfig()

	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS section (name VARCHAR(255) PRIMARY KEY, state VARCHAR(255))`)
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/", Home)
	router.GET("/version", Version)
	router.ServeFiles("/client/*filepath", http.Dir("./client/"))
	router.GET("/section/:sectionName", GetSection)
	router.POST("/section/:sectionName", PostSection)

	log.Fatal(http.ListenAndServe(config.PortStr(), router))
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

	rows, err := db.Query(`SELECT state FROM section WHERE name=?`, sectionName)
	if err != nil {
		log.Fatal(err)
	}

	if !rows.Next() {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Not Found")
		return
	}

	var state string
	err = rows.Scan(&state)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, state)
	fmt.Println("Get section:", state)
}

func PostSection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sectionName := ps.ByName("sectionName")
	state := r.PostFormValue("state")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	rows, err := tx.Query(`SELECT state FROM section WHERE name=?`, sectionName)
	if err != nil {
		log.Fatal(err)
	}

	if rows.Next() {
		var oldState string
		err = rows.Scan(&oldState)
		if err != nil {
			log.Fatal(err)
		}

		if oldState != "" && state != "" {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintln(w, "Conflict")
			tx.Rollback()
			return
		}

		_, err = tx.Exec(`UPDATE section SET state=? WHERE name=?`, state, sectionName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Update failed")
			log.Println(err)
			tx.Rollback()
			return
		}
	} else {
		_, err = tx.Exec(`INSERT INTO section (name, state) VALUES (?, ?)`, sectionName, state)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Update failed")
			log.Println(err)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Update failed")
		log.Println(err)
		tx.Rollback()
	}

	fmt.Printf("Post section: section[%s] = %s\n", sectionName, state)
}

type Config struct {
	Server struct {
		Port int `toml:"port"`
	} `toml:"server"`
}

var config Config

func LodaConfig() {
	_, err := toml.DecodeFile("config.tml", &config)
	if err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) PortStr() string {
	if cfg.Server.Port == 80 {
		return ""
	} else {
		return fmt.Sprintf(":%d", cfg.Server.Port)
	}
}
