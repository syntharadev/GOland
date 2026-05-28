package db

import (
	"testing"
)

// TestSQLiteCRUD verifica el funcionamiento y la resistencia a inyecciones SQL
func TestSQLiteCRUD(t *testing.T) {
	// Usamos una base de datos en memoria para no afectar el archivo de producción
	repo, err := InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Error inicializando DB en memoria: %v", err)
	}
	defer repo.Close()

	// 1. Test de Guardado Inicial (Create)
	t.Run("Crear Usuario Nuevo", func(t *testing.T) {
		err := repo.SaveProgress("GopherTest", "Ciberseguridad", "Hacking", 1)
		if err != nil {
			t.Errorf("Fallo al guardar progreso: %v", err)
		}
	})

	// 2. Test de Recuperación (Read)
	t.Run("Recuperar Usuario Existente", func(t *testing.T) {
		user, err := repo.GetUserProgress("GopherTest")
		if err != nil {
			t.Errorf("Fallo al recuperar usuario: %v", err)
		}
		if user.NivelActual != 1 {
			t.Errorf("Se esperaba Nivel 1, se obtuvo %d", user.NivelActual)
		}
	})

	// 3. Test de Actualización (Update/Upsert)
	t.Run("Actualizar Nivel de Usuario", func(t *testing.T) {
		err := repo.SaveProgress("GopherTest", "Ciberseguridad", "Hacking", 5)
		if err != nil {
			t.Errorf("Fallo al actualizar progreso: %v", err)
		}
		user, _ := repo.GetUserProgress("GopherTest")
		if user.NivelActual != 5 {
			t.Errorf("Upsert fallido. Se esperaba Nivel 5, se obtuvo %d", user.NivelActual)
		}
	})

	// 4. Test de Seguridad (SQL Injection Prevent)
	t.Run("Resistencia SQL Injection", func(t *testing.T) {
		maliciousNick := "GopherTest' OR '1'='1"
		// Intentamos guardar un Nick malicioso
		repo.SaveProgress(maliciousNick, "Test", "Test", 1)
		
		// Si el driver es seguro, lo guardará como un string literal y no alterará la tabla
		user, err := repo.GetUserProgress(maliciousNick)
		if err != nil || user.Nick != maliciousNick {
			t.Errorf("Vulnerabilidad detectada o fallo de escape de caracteres: %v", err)
		}
	})
}
