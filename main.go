package main

import (
	"html/template"
	"log"
	"net/http"

	"groupie-tracker/auth"
	"groupie-tracker/database"
	"groupie-tracker/room"
)

func main() {
	err := database.InitDB("groupie_tracker.db")
	if err != nil {
		log.Fatal("Erreur initialisation BDD:", err)
	}
	defer database.CloseDB()

	hub := room.NewHub()
	go hub.Run()

	setupRoutes(hub)

	log.Println("Serveur demarre sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes(hub *room.Hub) {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/register", auth.RegisterHandler(database.DB))
	http.HandleFunc("/login", auth.LoginHandler(database.DB))
	http.HandleFunc("/logout", auth.LogoutHandler(database.DB))

	http.HandleFunc("/", auth.AuthMiddleware(database.DB, landingPageHandler))
	http.HandleFunc("/room/create", auth.AuthMiddleware(database.DB, createRoomHandler))
	http.HandleFunc("/room/join", auth.AuthMiddleware(database.DB, joinRoomHandler))
	http.HandleFunc("/room/", auth.AuthMiddleware(database.DB, roomHandler))

	http.HandleFunc("/ws", auth.AuthMiddleware(database.DB, func(w http.ResponseWriter, r *http.Request) {
		websocketHandler(hub, w, r)
	}))
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	pseudo := auth.GetUserPseudo(r)

	data := struct {
		Pseudo string
	}{
		Pseudo: pseudo,
	}

	tmpl := template.Must(template.ParseFiles("templates/landing.html"))
	tmpl.Execute(w, data)
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	gameType := r.FormValue("game_type")

	newRoom, err := room.CreateRoom(database.DB, gameType, userID)
	if err != nil {
		http.Error(w, "Erreur creation salle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/room/"+newRoom.Code, http.StatusSeeOther)
}

func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Methode non autorisee", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	roomCode := r.FormValue("room_code")

	err := room.JoinRoom(database.DB, roomCode, userID)
	if err != nil {
		http.Error(w, "Erreur: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/room/"+roomCode, http.StatusSeeOther)
}

func roomHandler(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Path[len("/room/"):]

	currentRoom, err := room.GetRoomByCode(database.DB, roomCode)
	if err != nil {
		http.Error(w, "Salle introuvable", http.StatusNotFound)
		return
	}

	userID := auth.GetUserID(r)
	isInRoom := false
	for _, player := range currentRoom.Players {
		if player.UserID == userID {
			isInRoom = true
			break
		}
	}

	if !isInRoom {
		http.Error(w, "Vous n'etes pas dans cette salle", http.StatusForbidden)
		return
	}

	data := struct {
		Room    *room.Room
		IsReady bool
		UserID  int
	}{
		Room:    currentRoom,
		IsReady: room.IsRoomReady(*currentRoom),
		UserID:  userID,
	}

	if currentRoom.GameType == "blindtest" {
		tmpl := template.Must(template.ParseFiles("templates/blindtest.html"))
		tmpl.Execute(w, data)
	} else {
		tmpl := template.Must(template.ParseFiles("templates/petitbac.html"))
		tmpl.Execute(w, data)
	}
}

func websocketHandler(hub *room.Hub, w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	pseudo := auth.GetUserPseudo(r)

	roomCode := r.URL.Query().Get("room")
	currentRoom, err := room.GetRoomByCode(database.DB, roomCode)
	if err != nil {
		http.Error(w, "Salle introuvable", http.StatusNotFound)
		return
	}

	room.ServeWS(hub, w, r, currentRoom.ID, userID, pseudo)
}
