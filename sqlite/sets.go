package sqlite

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

// SetsInit creates 'sets' table and indexes
func (s *Service) setsInit() error {
	log.Println("Initing sets table")

	_, err := s.connection.Exec(`
		CREATE TABLE IF NOT EXISTS sets (
			id text not null primary key,
			name text
		)
	`)
	if err != nil {
		return errors.Wrap(err, "can't create sets table")
	}

	_, err = s.connection.Exec("CREATE UNIQUE INDEX IF NOT EXISTS name ON sets (name)")
	if err != nil {
		return errors.Wrap(err, "can't create index ON sets (name)")
	}

	return nil
}

// SetsInsert inserts new set
func (s *Service) SetsInsert(id, name string) error {
	stmt, err := s.connection.Prepare("INSERT INTO sets(id, name) VALUES(?, ?)")
	if err != nil {
		return errors.Wrap(err, "can't prepare query")
	}
	_, err = stmt.Exec(id, name)
	if err != nil {
		return errors.Wrapf(err, "Can't insert set '%s'", name)
	}
	return nil
}

// SetsGetIDByName returns set id by name
func (s *Service) SetsGetIDByName(name string) (string, error) {
	var setID string
	err := s.connection.QueryRow("SELECT id FROM sets WHERE name=?", name).Scan(&setID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", errors.Wrap(err, "can't select")
	}

	return setID, nil
}
