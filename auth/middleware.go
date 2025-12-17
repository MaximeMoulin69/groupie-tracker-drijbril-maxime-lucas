package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserPseudoKey contextKey = "userPseudo"

func AuthMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := ValidateSession(db, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
		ctx = context.WithValue(ctx, UserPseudoKey, user.Pseudo)
		
		next(w, r.WithContext(ctx))
	}
}

func GetUserID(r *http.Request) int {
	userID, ok := r.Context().Value(UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}

func GetUserPseudo(r *http.Request) string {
	pseudo, ok := r.Context().Value(UserPseudoKey).(string)
	if !ok {
		return ""
	}
	return pseudo
}

func GetUserIDStr(r *http.Request) string {
	return strconv.Itoa(GetUserID(r))
}