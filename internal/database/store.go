package database

import (
	"database/sql"
	"gotube/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type DB struct {
	conn *sql.DB
}

func InitDB() (*DB, error) {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".config", "gotube")
	os.MkdirAll(dbPath, 0755)
	
	db, err := sql.Open("sqlite3", filepath.Join(dbPath, "data.db"))
	if err != nil { return nil, err }

	createTables := `
	CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT);
	CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, title TEXT, url TEXT, path TEXT, timestamp INTEGER);
	`
	_, err = db.Exec(createTables)
	return &DB{conn: db}, err
}

func (d *DB) SaveHistory(title, url, path string) {
	d.conn.Exec("INSERT INTO history (title, url, path, timestamp) VALUES (?, ?, ?, ?)", title, url, path, 0)
}

func (d *DB) GetHistory() []models.HistoryEntry {
	rows, err := d.conn.Query("SELECT id, COALESCE(title, 'Unknown Video'), url, path FROM history ORDER BY id DESC LIMIT 50")
	if err != nil { return []models.HistoryEntry{} }
	defer rows.Close()

	var history []models.HistoryEntry
	for rows.Next() {
		var h models.HistoryEntry
		rows.Scan(&h.ID, &h.Title, &h.URL, &h.FilePath)
		history = append(history, h)
	}
	return history
}

func (d *DB) SaveSetting(key, value string) error {
	_, err := d.conn.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

func (d *DB) GetSetting(key string) string {
	var value string
	err := d.conn.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil { return "" }
	return value
}

func (d *DB) LoadSettings() models.AppSettings {
	return models.AppSettings{
		LastSavePath: d.GetSetting("LastSavePath"),
		ClientSpoof:  d.GetSetting("ClientSpoof"),
		CookiesPath:  d.GetSetting("CookiesPath"),
		Language:     d.GetSetting("Language"),
	}
}
