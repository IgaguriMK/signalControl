package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

const VERSION = "0.1.0"

var db *sql.DB

func main() {
	logf, err := os.OpenFile("mcsignal.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logf)
	log.SetFlags(log.Lshortfile)

	log.Println("---- launch ----")

	LoadConfig()

	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS world (id INTEGER PRIMARY KEY, name VARCHAR(255))`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS section (id INTEGER PRIMARY KEY, world_id INTEGER, name VARCHAR(255), state VARCHAR(255))`)
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/", Home)
	router.GET("/version", Version)
	router.ServeFiles("/client/*filepath", http.Dir("./client/"))
	router.GET("/world/:worldName/section/:sectionName", GetSection)
	router.POST("/world/:worldName/section/:sectionName", PostSection)

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
	worldName := ps.ByName("worldName")
	sectionName := ps.ByName("sectionName")

	rows, err := db.Query(
		`SELECT s.state
			FROM section s
			INNER JOIN world w
			ON s.world_id = w.id
			WHERE w.name=? AND s.name=?`,
		worldName,
		sectionName,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
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
	worldName := ps.ByName("worldName")
	sectionName := ps.ByName("sectionName")
	state := r.PostFormValue("state")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	worlds, err := tx.Query(`SELECT id FROM world WHERE name=?`, worldName)
	if err != nil {
		log.Fatal(err)
	}
	defer worlds.Close()
	if !worlds.Next() {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Not Found")
		return
	}

	var worldId int
	err = worlds.Scan(&worldId)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := tx.Query(`SELECT state FROM section WHERE world_id=? AND name=? `, worldId, sectionName)
	if err != nil {
		log.Fatal(err)
	}

	hasSection := rows.Next()
	if hasSection {
		var oldState string
		err = rows.Scan(&oldState)
		if err != nil {
			log.Fatal(err)
		}
		rows.Close()

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
		rows.Close()
		_, err = tx.Exec(`INSERT INTO section (world_id, name, state) VALUES (?, ?, ?)`, worldId, sectionName, state)
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

	fmt.Fprint(w, "Success")

	fmt.Printf("Post section: section[%s] = %s\n", sectionName, state)
}

type Config struct {
	Server struct {
		Port int `toml:"port"`
	} `toml:"server"`
}

var config Config

func LoadConfig() {
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
