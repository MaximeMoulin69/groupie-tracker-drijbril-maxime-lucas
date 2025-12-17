package auth

import (
	"database/sql"
	"html/template"
	"net/http"
)

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl := template.Must(template.ParseFiles("templates/login.html"))
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == "POST" {
			identifier := r.FormValue("identifier")
			password := r.FormValue("password")

			var user User
			err := db.QueryRow(`
				SELECT id, pseudo, email, password_hash 
				FROM users 
				WHERE pseudo = ? OR email = ?
			`, identifier, identifier).Scan(&user.ID, &user.Pseudo, &user.Email, &user.PasswordHash)

			if err != nil || !CheckPassword(password, user.PasswordHash) {
				http.Error(w, "Identifiants incorrects", http.StatusUnauthorized)
				return
			}

			token, err := CreateSession(db, user.ID)
			if err != nil {
				http.Error(w, "Erreur serveur", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}