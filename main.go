package main

import (
    "os"
    "database/sql"
	"encoding/json"
    "log"
    "net/http"
    "bytes"

    "github.com/joho/godotenv"
    _ "github.com/mattn/go-sqlite3"
)

type Link struct {
    ID          int    `json:"id"`
	URL         string `json:"url"`
}

var database *sql.DB

func main() {
	var err error

    err = godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

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
    http.HandleFunc("/add-link", addLink)

    log.Fatal(http.ListenAndServe(":8080", nil))
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


func addLink(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    url := r.FormValue("url")
    if url == "" {
        http.Error(w, "URL is required", http.StatusBadRequest)
        return
    }

    webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
    if webhookURL == "" {
        log.Println("Discord webhook URL is not set")
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    jsonStr := []byte(`{"content": "[_](` + url + `)", "username": "Synaptic Feed"}`)
    req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonStr))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Error sending message to Discord", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    statement, err := database.Prepare("INSERT INTO links (url) VALUES (?)")
    if err != nil {
        http.Error(w, "Error preparing insert statement", http.StatusInternalServerError)
        return
    }
    defer statement.Close()

    _, err = statement.Exec(url)
    if err != nil {
        http.Error(w, "Error executing insert statement", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Link added successfully"})
}
