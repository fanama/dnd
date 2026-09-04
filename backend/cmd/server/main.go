package main

import (
	"encoding/json"
	"log"
	"net/http"

	"dnd-backend/internal/repository"
	"dnd-backend/internal/services"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

func main() {
	// Dependency Injection
	repo, err := repository.NewSQLiteRepository("game.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer repo.Close()

	gm := services.NewGameManager(repo)

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		pseudo := r.URL.Path[len("/ws/"):]
		if pseudo == "" {
			http.Error(w, "Pseudo required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		// Wait for initial character info
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read initial msg error:", err)
			return
		}

		var charInfo map[string]string
		json.Unmarshal(msg, &charInfo)

		gm.Connect(pseudo, conn, charInfo)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				gm.Disconnect(pseudo)
				break
			}

			var action services.Action
			if err := json.Unmarshal(msg, &action); err == nil {
				gm.HandleAction(pseudo, action)
			}
		}
	})

	log.Println("Server starting on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
