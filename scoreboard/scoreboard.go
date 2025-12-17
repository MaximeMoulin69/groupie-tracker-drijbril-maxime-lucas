package scoreboard

import (
	"database/sql"
)

type ScoreboardEntry struct {
	Pseudo string
	Score  int
}

func GetGameScoreboard(db *sql.DB, roomID int, gameType string) ([]ScoreboardEntry, error) {
	rows, err := db.Query(`
		SELECT u.pseudo, SUM(s.score) as total_score
		FROM scores s
		JOIN users u ON s.user_id = u.id
		WHERE s.room_id = ? AND s.game_type = ?
		GROUP BY u.id, u.pseudo
		ORDER BY total_score DESC
	`, roomID, gameType)

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

func GetRoundScoreboard(db *sql.DB, roomID int, gameType string, roundNumber int) ([]ScoreboardEntry, error) {
	rows, err := db.Query(`
		SELECT u.pseudo, s.score
		FROM scores s
		JOIN users u ON s.user_id = u.id
		WHERE s.room_id = ? AND s.game_type = ? AND s.round_number = ?
		ORDER BY s.score DESC
	`, roomID, gameType, roundNumber)

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

func GetPlayerTotalScore(db *sql.DB, roomID int, userID int) (int, error) {
	var totalScore int

	err := db.QueryRow(`
		SELECT COALESCE(SUM(score), 0) as total
		FROM scores
		WHERE room_id = ? AND user_id = ?
	`, roomID, userID).Scan(&totalScore)

	return totalScore, err
}