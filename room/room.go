package room

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type Room struct {
	ID         int
	Code       string
	GameType   string
	HostID     int
	MaxPlayers int
	Status     string
	CreatedAt  time.Time
	Players    []Player
}

type Player struct {
	UserID   int
	Pseudo   string
	JoinedAt time.Time
}

func GenerateRoomCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateRoom(db *sql.DB, gameType string, hostID int) (*Room, error) {
	if gameType != "blindtest" && gameType != "petitbac" {
		return nil, errors.New("type de jeu invalide")
	}

	code, err := GenerateRoomCode()
	if err != nil {
		return nil, err
	}

	result, err := db.Exec(
		"INSERT INTO rooms (code, game_type, host_id, status) VALUES (?, ?, ?, 'waiting')",
		code, gameType, hostID,
	)
	if err != nil {
		return nil, err
	}

	roomID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		"INSERT INTO room_players (room_id, user_id) VALUES (?, ?)",
		roomID, hostID,
	)
	if err != nil {
		return nil, err
	}

	room := &Room{
		ID:         int(roomID),
		Code:       code,
		GameType:   gameType,
		HostID:     hostID,
		MaxPlayers: 10,
		Status:     "waiting",
		CreatedAt:  time.Now(),
	}

	return room, nil
}

func GetRoomByCode(db *sql.DB, code string) (*Room, error) {
	var room Room

	err := db.QueryRow(`
		SELECT id, code, game_type, host_id, max_players, status, created_at
		FROM rooms
		WHERE code = ?
	`, code).Scan(&room.ID, &room.Code, &room.GameType, &room.HostID, &room.MaxPlayers, &room.Status, &room.CreatedAt)

	if err != nil {
		return nil, errors.New("salle introuvable")
	}

	room.Players, err = GetRoomPlayers(db, room.ID)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func JoinRoom(db *sql.DB, roomCode string, userID int) error {
	room, err := GetRoomByCode(db, roomCode)
	if err != nil {
		return err
	}

	if len(room.Players) >= room.MaxPlayers {
		return errors.New("la salle est pleine")
	}

	if room.Status != "waiting" {
		return errors.New("la partie a deja commence")
	}

	_, err = db.Exec(
		"INSERT INTO room_players (room_id, user_id) VALUES (?, ?)",
		room.ID, userID,
	)

	if err != nil {
		return errors.New("vous etes deja dans cette salle")
	}

	return nil
}

func GetRoomPlayers(db *sql.DB, roomID int) ([]Player, error) {
	rows, err := db.Query(`
		SELECT u.id, u.pseudo, rp.joined_at
		FROM room_players rp
		JOIN users u ON rp.user_id = u.id
		WHERE rp.room_id = ?
		ORDER BY rp.joined_at ASC
	`, roomID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		err := rows.Scan(&player.UserID, &player.Pseudo, &player.JoinedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func IsRoomReady(r Room) bool {
	if len(r.Players) < 2 {
		return false
	}

	if r.Status != "waiting" {
		return false
	}

	return true
}

func StartGame(db *sql.DB, roomID int, hostID int) error {
	var currentHostID int
	err := db.QueryRow("SELECT host_id FROM rooms WHERE id = ?", roomID).Scan(&currentHostID)
	if err != nil {
		return errors.New("salle introuvable")
	}

	if currentHostID != hostID {
		return errors.New("seul l'hote peut demarrer la partie")
	}

	_, err = db.Exec("UPDATE rooms SET status = 'playing' WHERE id = ?", roomID)
	return err
}

func LeaveRoom(db *sql.DB, roomID int, userID int) error {
	_, err := db.Exec(
		"DELETE FROM room_players WHERE room_id = ? AND user_id = ?",
		roomID, userID,
	)
	return err
}