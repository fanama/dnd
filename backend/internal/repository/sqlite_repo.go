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
		stats TEXT,
		equipement TEXT,
		quests TEXT
	);`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	for _, col := range []string{"equipement", "quests", "or", "xp"} {
		if err := ensureColumn(db, "characters", col); err != nil {
			return err
		}
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS world_state (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		npcs TEXT,
		locations TEXT
	);`); err != nil {
		return err
	}
	return nil
}

func ensureColumn(db *sql.DB, table, column string) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('`+table+`') WHERE name = ?`, column)
	if err != nil {
		return err
	}
	exists := rows.Next()
	rows.Close()
	if exists {
		return nil
	}
	// Quoted because "or" is a reserved keyword in SQLite.
	_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + quoteIdent(column) + ` TEXT`)
	return err
}

func quoteIdent(name string) string {
	return `"` + name + `"`
}

func (r *SQLiteRepository) SaveCharacter(pseudo, nom, classe, lieu string, pv, maxPv int, inventaire, stats, equipement, quests string, or, xp float64) error {
	query := `INSERT OR REPLACE INTO characters (pseudo, nom, classe, lieu, pv, max_pv, inventaire, stats, equipement, quests, "or", xp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, pseudo, nom, classe, lieu, pv, maxPv, inventaire, stats, equipement, quests, or, xp)
	return err
}

func (r *SQLiteRepository) GetCharacter(pseudo string) (nom, classe, lieu string, pv, maxPv int, or, xp float64, inventaire, stats, equipement, quests string, err error) {
	query := `SELECT nom, classe, lieu, pv, max_pv, COALESCE("or", 0), COALESCE(xp, 0), inventaire, stats, equipement, COALESCE(quests, '') FROM characters WHERE pseudo = ?`
	err = r.db.QueryRow(query, pseudo).Scan(&nom, &classe, &lieu, &pv, &maxPv, &or, &xp, &inventaire, &stats, &equipement, &quests)
	return
}

// Checkpoint flushes the SQLite WAL into the main database file. It is called
// before a backup so the copy captures the latest writes.
func (r *SQLiteRepository) Checkpoint() {
	r.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
}

func (r *SQLiteRepository) DeleteCharacter(pseudo string) error {
	query := `DELETE FROM characters WHERE pseudo = ?`
	_, err := r.db.Exec(query, pseudo)
	return err
}

func (r *SQLiteRepository) GetAllCharacters() (map[string]struct {
	Nom, Classe, Lieu, Inventaire, Stats, Equipement, Quests string
	PV, MaxPV                                                int
	Or, Xp                                                   float64
}, error) {
	query := `SELECT pseudo, nom, classe, lieu, pv, max_pv, inventaire, stats, equipement, COALESCE(quests, ''), COALESCE("or", 0), COALESCE(xp, 0) FROM characters`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]struct {
		Nom, Classe, Lieu, Inventaire, Stats, Equipement, Quests string
		PV, MaxPV                                                int
		Or, Xp                                                   float64
	})
	for rows.Next() {
		var pseudo, nom, classe, lieu, inventaire, stats, equipement, quests string
		var pv, maxPv int
		var or, xp float64
		if err := rows.Scan(&pseudo, &nom, &classe, &lieu, &pv, &maxPv, &inventaire, &stats, &equipement, &quests, &or, &xp); err != nil {
			continue
		}
		result[pseudo] = struct {
			Nom, Classe, Lieu, Inventaire, Stats, Equipement, Quests string
			PV, MaxPV                                                int
			Or, Xp                                                   float64
		}{
			Nom: nom, Classe: classe, Lieu: lieu, Inventaire: inventaire, Stats: stats, Equipement: equipement, Quests: quests, PV: pv, MaxPV: maxPv, Or: or, Xp: xp,
		}
	}
	return result, nil
}

func (r *SQLiteRepository) SaveWorld(npcs, locations string) error {
	query := `INSERT OR REPLACE INTO world_state (id, npcs, locations) VALUES (1, ?, ?)`
	_, err := r.db.Exec(query, npcs, locations)
	return err
}

func (r *SQLiteRepository) LoadWorld() (npcs, locations string, err error) {
	query := `SELECT npcs, locations FROM world_state WHERE id = 1`
	err = r.db.QueryRow(query).Scan(&npcs, &locations)
	return
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
