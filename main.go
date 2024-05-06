package main

import (
	"os"
	"log"
	"net/http"
	"gopoc/handlers"
	"gopoc/websocket"
	"gopoc/db"
	"html/template"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func addCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.Connect(os.Getenv("DATABASE_URL"))

	templates, errr := template.ParseFiles("public/test/index_raw.html")
    if errr != nil {
		log.Println("Failed to parse templates")
        panic(errr)
    }

	r := mux.NewRouter()

	r.Handle("/public/test/", http.StripPrefix("/public/test/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data := struct {
            Title string
        }{
            Title: "Home Page",
        }
        err := templates.ExecuteTemplate(w, "index_raw.html", data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

	// WebSocket route
	r.HandleFunc("/ws/{channelID}", websocket.HandleConnections)

	// REST API routes
	r.HandleFunc("/signup", handlers.CreateUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", handlers.GetUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/users", handlers.ListUsers).Methods("GET", "OPTIONS")
	r.HandleFunc("/messages/{channelID}", handlers.GetMessages).Methods("GET", "OPTIONS")
	r.HandleFunc("/channels", handlers.CreateChannel).Methods("POST", "OPTIONS")
	r.HandleFunc("/channels", handlers.GetChannels).Methods("GET", "OPTIONS")

	port := os.Getenv("PORT")

	log.Println("Starting server on port", port)
	err = http.ListenAndServe(port, addCORS(r))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
