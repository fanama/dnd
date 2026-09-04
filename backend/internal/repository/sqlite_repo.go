package repository

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS characters (
		pseudo TEXT PRIMARY KEY,
		nom TEXT,
		classe TEXT,
		pv INTEGER,
		max_pv INTEGER,
		lieu TEXT,
		inventaire TEXT,
		stats TEXT
	);`
	_, err := db.Exec(query)
	return err
}

func (r *SQLiteRepository) SaveCharacter(pseudo, nom, classe, lieu string, pv, maxPv int, inventaire string, stats string) error {
	query := \`INSERT OR REPLACE INTO characters (pseudo, nom, classe, pv, max_pv, lieu, inventaire, stats) VALUES (?, ?, ?, ?, ?, ?, ?, ?)\`
	_, err := r.db.Exec(query, pseudo, nom, classe, lieu, pv, maxPv, inventaire, stats)
	return err
}

func (r *SQLiteRepository) GetCharacter(pseudo string) (nom, classe, lieu string, pv, maxPv int, inventaire, stats string, err error) {
	query := \`SELECT nom, classe, lieu, pv, max_pv, inventaire, stats FROM characters WHERE pseudo = ?\`
	err = r.db.QueryRow(query, pseudo).Scan(&nom, &classe, &lieu, &pv, &maxPv, &inventaire, &stats)
	return
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
