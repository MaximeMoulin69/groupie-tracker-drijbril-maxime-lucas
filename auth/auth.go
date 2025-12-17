package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Pseudo       string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func ValidatePseudo(pseudo string) error {
	if len(pseudo) == 0 {
		return errors.New("le pseudo ne peut pas etre vide")
	}

	if !unicode.IsUpper(rune(pseudo[0])) {
		return errors.New("le pseudo doit commencer par une majuscule")
	}

	return nil
}

func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("format d'email invalide")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("le mot de passe doit contenir au moins 12 caracteres")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("le mot de passe doit contenir au moins une majuscule")
	}
	if !hasLower {
		return errors.New("le mot de passe doit contenir au moins une minuscule")
	}
	if !hasNumber {
		return errors.New("le mot de passe doit contenir au moins un chiffre")
	}
	if !hasSpecial {
		return errors.New("le mot de passe doit contenir au moins un caractere special")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateSession(db *sql.DB, userID int) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = db.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)

	return token, err
}

func ValidateSession(db *sql.DB, token string) (*User, error) {
	var user User
	var expiresAt time.Time

	err := db.QueryRow(`
		SELECT u.id, u.pseudo, u.email, u.password_hash, u.created_at, s.expires_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.session_token = ?
	`, token).Scan(&user.ID, &user.Pseudo, &user.Email, &user.PasswordHash, &user.CreatedAt, &expiresAt)

	if err != nil {
		return nil, errors.New("session invalide")
	}

	if time.Now().After(expiresAt) {
		db.Exec("DELETE FROM sessions WHERE session_token = ?", token)
		return nil, errors.New("session expiree")
	}

	return &user, nil
}

func DeleteSession(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE session_token = ?", token)
	return err
}
