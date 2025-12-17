package game

import (
	"database/sql"
	"errors"
)

type BlindTestConfig struct {
	ID           int
	RoomID       int
	Playlist     string
	ResponseTime int
	NbrRounds    int
}

type ScoreboardEntry struct {
	Pseudo string
	Score  int
}

func CreateBlindTestConfig(db *sql.DB, roomID int, playlist string, responseTime int, nbrRounds int) error {
	validPlaylists := map[string]bool{"Rock": true, "Rap": true, "Pop": true}
	if !validPlaylists[playlist] {
		return errors.New("playlist invalide. Choisis une des playlist ! (Rock, Rap, Pop)")
	}

	if responseTime == 0 {
		responseTime = 37
	}

	_, err := db.Exec(`
		INSERT INTO blindtest_config (room_id, playlist, response_time, nbr_rounds)
		VALUES (?, ?, ?, ?)
	`, roomID, playlist, responseTime, nbrRounds)

	return err
}


func CalculateBlindTestPoints(position int, totalPlayers int) int {
	basePoints := 100

	if position == 1 {
		return basePoints
	}
	if position == 2 {
		return 75
	}
	if position == 3 {
		return 50
	}

	points := basePoints - (position-1)*15
	if points < 10 {
		points = 10
	}

	return points
}

//pas pu finir