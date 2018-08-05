package sqlite

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // sqlite3
	"github.com/pkg/errors"
)

// Service это сервис для работы с sqlite
type Service struct {
	connection *sql.DB
}

// NewService создаёт новый сервис для работы с sqlite
func NewService(dbFile string) (*Service, error) {

	service := Service{}

	log.Println("Opening DB")
	connection, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, errors.Wrap(err, "can't open DB")
	}
	service.connection = connection

	err = service.photosInit()
	if err != nil {
		return nil, errors.Wrap(err, "can't init photos table")
	}

	err = service.setsInit()
	if err != nil {
		return nil, errors.Wrap(err, "can't init sets table")
	}

	return &service, nil
}
