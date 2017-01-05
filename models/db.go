package models

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // sqlite3
	"log"
)

var db *sql.DB

// InitDb inits dababase
func InitDb(dbFile string) {
	var err error

	log.Println("Opening DB")
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalln(err)
	}

	err = PhotosInit()
	if err!=nil {
		log.Fatalln(err)
	}

	err = SetsInit()
	if err!=nil {
		log.Fatalln(err)
	}
}



