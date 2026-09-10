package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"dnd-backend/internal/repository"
	"dnd-backend/internal/services"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// spaHandler serves a built Svelte app (dist folder) and falls back to index.html.
func spaHandler(dist string) http.Handler {
	fs := http.FileServer(http.Dir(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		p := filepath.Join(dist, filepath.FromSlash(r.URL.Path))
		if r.URL.Path != "/" {
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})
}

func main() {
	// Dependency Injection
	repo, err := repository.NewSQLiteRepository(envOr("DB_PATH", "game.db"))
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer repo.Close()

	gm := services.NewGameManager(repo)
	defer gm.Close()

	// Option A: one process serves static apps and the WebSocket for both.
	playerDist := envOr("PLAYER_DIST", "../frontend/dist")
	dmDist := envOr("DM_DIST", "../dm-frontend/dist")

	http.Handle("/dm/", http.StripPrefix("/dm/", spaHandler(dmDist)))
	http.Handle("/", spaHandler(playerDist))
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

	port := envOr("PORT", "8000")
	log.Println("Server starting on :" + port)

	srv := &http.Server{Addr: ":" + port, Handler: nil}
	done := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		gm.DisconnectAll()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		close(done)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe:", err)
	}
	<-done
}
