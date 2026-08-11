package smartcontact

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"migrated-app/internal/config"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("config load error: %v", err)
		return
	}
	d, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Printf("db open error: %v", err)
		return
	}
	if err := d.Ping(); err != nil {
		log.Printf("db ping error: %v", err)
		return
	}
	db = d
	if err := createTable(); err != nil {
		log.Printf("create table error: %v", err)
	}
}

func createTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS contacts (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT,
		phone TEXT
	)`)
	return err
}

type Contact struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func BuildRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			switch r.Method {
			case http.MethodGet:
				listContacts(w, r)
			case http.MethodPost:
				createContact(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		// /contacts/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) == 1 {
			id, err := strconv.Atoi(parts[0])
			if err != nil {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				getContact(w, r, id)
			case http.MethodPut:
				updateContact(w, r, id)
			case http.MethodDelete:
				deleteContact(w, r, id)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listContacts(w, r)
		case http.MethodPost:
			createContact(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/contacts/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getContact(w, r, id)
		case http.MethodPut:
			updateContact(w, r, id)
		case http.MethodDelete:
			deleteContact(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if db == nil {
			http.Error(w, `{"status":"db unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		if err := db.Ping(); err != nil {
			http.Error(w, `{"status":"db unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func listContacts(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	rows, err := db.Query("SELECT id, name, email, phone FROM contacts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	contacts := []Contact{}
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		contacts = append(contacts, c)
	}
	writeJSON(w, http.StatusOK, contacts)
}

func createContact(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	var c Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := db.QueryRow(
		"INSERT INTO contacts (name, email, phone) VALUES ($1, $2, $3) RETURNING id",
		c.Name, c.Email, c.Phone,
	).Scan(&c.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func getContact(w http.ResponseWriter, r *http.Request, id int) {
	if db == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	var c Contact
	err := db.QueryRow("SELECT id, name, email, phone FROM contacts WHERE id=$1", id).
		Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func updateContact(w http.ResponseWriter, r *http.Request, id int) {
	if db == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	var c Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.ID = id
	_, err := db.Exec(
		"UPDATE contacts SET name=$1, email=$2, phone=$3 WHERE id=$4",
		c.Name, c.Email, c.Phone, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func deleteContact(w http.ResponseWriter, r *http.Request, id int) {
	if db == nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	_, err := db.Exec("DELETE FROM contacts WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}