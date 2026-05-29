package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository struct {
	db *sql.DB
}

// InitDB inicializa la conexión con Supabase (PostgreSQL) usando el driver pgx
func InitDB(connStr string) (*Repository, error) {
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL no especificada. La base de datos Supabase/PostgreSQL es obligatoria")
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos Supabase: %w", err)
	}

	// Verificar conexión
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error al conectar con Supabase (Ping failed): %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS progreso_usuarios (
		nick VARCHAR(255) PRIMARY KEY,
		dominio TEXT NOT NULL,
		objetivo TEXT NOT NULL,
		nivel_actual INTEGER NOT NULL
	);`

	if _, err := db.Exec(query); err != nil {
		db.Close()
		return nil, fmt.Errorf("error al crear tabla progreso_usuarios en Supabase: %w", err)
	}

	log.Println("🗄️ Base de datos Supabase (PostgreSQL/pgx) inicializada correctamente en la nube.")
	return &Repository{db: db}, nil
}

// SaveProgress guarda o actualiza el progreso en Supabase con sintaxis nativa de Postgres
func (r *Repository) SaveProgress(nick, dominio, objetivo string, nivel int) error {
	query := `
	INSERT INTO progreso_usuarios (nick, dominio, objetivo, nivel_actual) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (nick) DO UPDATE SET 
		nivel_actual = EXCLUDED.nivel_actual;`

	_, err := r.db.Exec(query, nick, dominio, objetivo, nivel)
	if err != nil {
		return fmt.Errorf("error al guardar progreso en Supabase: %w", err)
	}
	return nil
}

// GetUserProgress recupera el progreso de un usuario en Supabase
func (r *Repository) GetUserProgress(nick string) (*UsuarioProgreso, error) {
	query := `SELECT nick, dominio, objetivo, nivel_actual FROM progreso_usuarios WHERE nick = $1`
	
	row := r.db.QueryRow(query, nick)
	
	var u UsuarioProgreso
	err := row.Scan(&u.Nick, &u.Dominio, &u.Objetivo, &u.NivelActual)
	if err != nil {
		return nil, err
	}
	
	return &u, nil
}

// Close cierra la conexión de Supabase
func (r *Repository) Close() {
	r.db.Close()
}

type UsuarioProgreso struct {
	Nick        string
	Dominio     string
	Objetivo    string
	NivelActual int
}
