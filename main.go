package main

import (
    "database/sql"
	"encoding/json"
    "log"
    "net/http"

    _ "github.com/mattn/go-sqlite3"
)

type Link struct {
    ID          int    `json:"id"`
	URL         string `json:"url"`
}

var database *sql.DB

func main() {
	var err error
    
	database, err = sql.Open("sqlite3", "./data.db")
    if err != nil {
        log.Fatal("Error opening database: ", err)
    }

    statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY, url TEXT)")
    if err != nil {
        log.Fatal("Error preparing database statement: ", err)
    }

    _, err = statement.Exec()
    if err != nil {
        log.Fatal("Error executing statement: ", err)
    }

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)


	statement.Exec()
    http.HandleFunc("/links", listLinks)

    log.Fatal(http.ListenAndServe(":8081", nil))
}


func listLinks(w http.ResponseWriter, r *http.Request) {
    var links []Link

    rows, err := database.Query("SELECT id, url FROM links")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var l Link
        if err := rows.Scan(&l.ID, &l.URL); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        links = append(links, l)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(links)
}
