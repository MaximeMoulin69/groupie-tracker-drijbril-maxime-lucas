package auth

import (
	"database/sql"
	"html/template"
	"net/http"
)

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl := template.Must(template.ParseFiles("templates/register.html"))
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == "POST" {
			pseudo := r.FormValue("pseudo")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirmPassword := r.FormValue("confirm_password")

			if err := ValidatePseudo(pseudo); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := ValidateEmail(email); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if password != confirmPassword {
				http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}

			if err := ValidatePassword(password); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			hashedPassword, err := HashPassword(password)
			if err != nil {
				http.Error(w, "Erreur serveur", http.StatusInternalServerError)
				return
			}

			_, err = db.Exec(
				"INSERT INTO users (pseudo, email, password_hash) VALUES (?, ?, ?)",
				pseudo, email, hashedPassword,
			)

			if err != nil {
				http.Error(w, "Pseudo ou email deja utilise", http.StatusConflict)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}