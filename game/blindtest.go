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

func GetBlindTestConfig(db *sql.DB, roomID int) (*BlindTestConfig, error) {
	var config BlindTestConfig

	err := db.QueryRow(`
		SELECT id, room_id, playlist, response_time, nbr_rounds
		FROM blindtest_config
		WHERE room_id = ?
	`, roomID).Scan(&config.ID, &config.RoomID, &config.Playlist, &config.ResponseTime, &config.NbrRounds)

	if err != nil {
		return nil, errors.New("configuration introuvable")
	}

	return &config, nil
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

func SaveBlindTestScore(db *sql.DB, roomID int, userID int, roundNumber int, scoreboardActualPointInGame int) error {
	_, err := db.Exec(`
		INSERT INTO scores (room_id, user_id, game_type, score, round_number)
		VALUES (?, ?, 'blindtest', ?, ?)
	`, roomID, userID, scoreboardActualPointInGame, roundNumber)

	return err
}

func GetBlindTestScoreboard(db *sql.DB, roomID int) ([]ScoreboardEntry, error) {
	rows, err := db.Query(`
		SELECT u.pseudo, SUM(s.score) as total_score
		FROM scores s
		JOIN users u ON s.user_id = u.id
		WHERE s.room_id = ? AND s.game_type = 'blindtest'
		GROUP BY u.id, u.pseudo
		ORDER BY total_score DESC
	`, roomID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scoreboard []ScoreboardEntry
	for rows.Next() {
		var entry ScoreboardEntry
		err := rows.Scan(&entry.Pseudo, &entry.Score)
		if err != nil {
			return nil, err
		}
		scoreboard = append(scoreboard, entry)
	}

	return scoreboard, nil
}