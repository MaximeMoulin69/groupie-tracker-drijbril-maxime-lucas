package game

import (
	"database/sql"
	"errors"
	"math/rand"
	"strings"
	"time"
)

const nbrs_manche = 9 // Variable constante pour le choix du nombre de manche

type PetitBacConfig struct {
	ID           int
	RoomID       int
	ResponseTime int
	NbrRounds    int
}

func CreatePetitBacConfig(db *sql.DB, roomID int, responseTime int, nbrRounds int) error {
	_, err := db.Exec(`
		INSERT INTO petitbac_config (room_id, response_time, nbr_rounds)
		VALUES (?, ?, ?)
	`, roomID, responseTime, nbrRounds)

	return err
}

func GetPetitBacConfig(db *sql.DB, roomID int) (*PetitBacConfig, error) {
	var config PetitBacConfig

	err := db.QueryRow(`
		SELECT id, room_id, response_time, nbr_rounds
		FROM petitbac_config
		WHERE room_id = ?
	`, roomID).Scan(&config.ID, &config.RoomID, &config.ResponseTime, &config.NbrRounds)

	if err != nil {
		return nil, errors.New("configuration introuvable")
	}

	return &config, nil
}

func AddCustomCategory(db *sql.DB, roomID int, categoryName string) error {
	_, err := db.Exec(`
		INSERT INTO petitbac_categories (room_id, category_name)
		VALUES (?, ?)
	`, roomID, categoryName)

	return err
}

func GetCustomCategories(db *sql.DB, roomID int) ([]string, error) {
	rows, err := db.Query(`
		SELECT category_name
		FROM petitbac_categories
		WHERE room_id = ?
	`, roomID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func UpdateCustomCategory(db *sql.DB, categoryID int, newName string) error {
	_, err := db.Exec(`
		UPDATE petitbac_categories
		SET category_name = ?
		WHERE id = ?
	`, newName, categoryID)

	return err
}

func DeleteCustomCategory(db *sql.DB, categoryID int) error {
	_, err := db.Exec(`
		DELETE FROM petitbac_categories
		WHERE id = ?
	`, categoryID)

	return err
}

func GenerateRandomLetter(usedLetters []string) string {
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVW"
	
	used := make(map[string]bool)
	for _, letter := range usedLetters {
		used[letter] = true
	}

	rand.Seed(time.Now().UnixNano())
	for {
		randomIndex := rand.Intn(len(alphabet))
		letter := string(alphabet[randomIndex])
		
		if !used[letter] {
			return letter
		}

		if len(used) >= len(alphabet) {
			return string(alphabet[randomIndex])
		}
	}
}

func ValidateAnswer(answer string, letter string) bool {
	if len(answer) == 0 {
		return false
	}
	return strings.ToUpper(string(answer[0])) == strings.ToUpper(letter)
}

func CalculatePetitBacPoints(validations int, totalVoters int, isUnique bool) int {
	requiredValidations := (totalVoters * 2) / 3
	isValid := validations >= requiredValidations

	if !isValid {
		return 0
	}

	if isUnique {
		return 2
	}

	return 1
}

func SavePetitBacScore(db *sql.DB, roomID int, userID int, roundNumber int, scoreboardActualPointInGame int) error {
	_, err := db.Exec(`
		INSERT INTO scores (room_id, user_id, game_type, score, round_number)
		VALUES (?, ?, 'petitbac', ?, ?)
	`, roomID, userID, scoreboardActualPointInGame, roundNumber)

	return err
}

func GetPetitBacScoreboard(db *sql.DB, roomID int) ([]ScoreboardEntry, error) {
	rows, err := db.Query(`
		SELECT u.pseudo, SUM(s.score) as total_score
		FROM scores s
		JOIN users u ON s.user_id = u.id
		WHERE s.room_id = ? AND s.game_type = 'petitbac'
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