package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	err = createTables()
	if err != nil {
		return err
	}

	log.Println("Base de donnees initialisee avec succes")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pseudo TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		game_type TEXT NOT NULL,
		host_id INTEGER NOT NULL,
		max_players INTEGER DEFAULT 10,
		status TEXT DEFAULT 'waiting',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (host_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS room_players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_id) REFERENCES rooms(id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(room_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS blindtest_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER UNIQUE NOT NULL,
		playlist TEXT NOT NULL,
		response_time INTEGER DEFAULT 37,
		nbr_rounds INTEGER NOT NULL,
		FOREIGN KEY (room_id) REFERENCES rooms(id)
	);

	CREATE TABLE IF NOT EXISTS petitbac_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER UNIQUE NOT NULL,
		response_time INTEGER NOT NULL,
		nbr_rounds INTEGER NOT NULL,
		FOREIGN KEY (room_id) REFERENCES rooms(id)
	);

	CREATE TABLE IF NOT EXISTS petitbac_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER NOT NULL,
		category_name TEXT NOT NULL,
		FOREIGN KEY (room_id) REFERENCES rooms(id)
	);

	CREATE TABLE IF NOT EXISTS scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		game_type TEXT NOT NULL,
		score INTEGER DEFAULT 0,
		round_number INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_id) REFERENCES rooms(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		session_token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(schema)
	return err
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Connexion a la base de donnees fermee")
	}
}
