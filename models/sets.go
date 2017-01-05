package models

import (
	"log"
	"fmt"
	"database/sql"
)

// SetsInit creates 'sets' table and indexes
func SetsInit() error {
	log.Println("Initing sets table")

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sets (
			id text not null primary key,
			name text
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS name ON sets (name)")
	if err != nil {
		return err
	}

	return nil
}

// SetsInsert inserts new set
func SetsInsert(id, name string) error {
	stmt, err := db.Prepare("INSERT INTO sets(id, name) VALUES(?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id, name)
	if err != nil {
		return fmt.Errorf("Can't insert set '%s' error:%s", name, err)
	}
	return nil
}

// SetsGetIDByName returns set id by name
func SetsGetIDByName(name string) (string, error) {
	var setID string
	err := db.QueryRow("SELECT id FROM sets WHERE name=?", name).
		Scan(&setID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return setID, nil
}