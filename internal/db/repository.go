package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func InitDB(filepath string) (*Repository, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS progreso_usuarios (
		nick TEXT PRIMARY KEY,
		dominio TEXT,
		objetivo TEXT,
		nivel_actual INTEGER
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	log.Println("🗄️ Base de datos SQLite inicializada correctamente.")
	return &Repository{db: db}, nil
}

func (r *Repository) SaveProgress(nick, dominio, objetivo string, nivel int) error {
	query := `
	INSERT INTO progreso_usuarios (nick, dominio, objetivo, nivel_actual) 
	VALUES (?, ?, ?, ?)
	ON CONFLICT(nick) DO UPDATE SET 
		nivel_actual=excluded.nivel_actual;`

	_, err := r.db.Exec(query, nick, dominio, objetivo, nivel)
	return err
}

func (r *Repository) Close() {
	r.db.Close()
}

type UsuarioProgreso struct {
	Nick        string
	Dominio     string
	Objetivo    string
	NivelActual int
}

// GetUserProgress busca a un usuario. Retorna error (sql.ErrNoRows) si no existe.
func (r *Repository) GetUserProgress(nick string) (*UsuarioProgreso, error) {
	query := `SELECT nick, dominio, objetivo, nivel_actual FROM progreso_usuarios WHERE nick = ?`
	
	row := r.db.QueryRow(query, nick)
	
	var u UsuarioProgreso
	err := row.Scan(&u.Nick, &u.Dominio, &u.Objetivo, &u.NivelActual)
	if err != nil {
		return nil, err
	}
	
	return &u, nil
}
