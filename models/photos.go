package models

import (
	"log"
	"fmt"
)

// PhotosInit creates 'photos' table and its indexes
func PhotosInit() error {
	log.Println("Initing photos table")

	// id has 'text' type (!)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS photos (
			id text not null primary key,
			path text not null,
			set_id text
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS fileindex ON photos (path)")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS setindex ON photos (set_id)")
	if err != nil {
		return err
	}

	return nil
}


// PhotosGetAll returns all photos in DB
func PhotosGetAll() (map[string]string, error)  {
	res := map[string]string{}

	rows, err := db.Query("SELECT path, id FROM photos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var path string
	var photoID string
	for rows.Next() {
		err := rows.Scan(&path, &photoID)
		if err != nil {
			return nil, fmt.Errorf("PhotosGetAll: %s", err)
		}
		res[path] = photoID
	}
	return res, nil
}

// PhotosInsert inserts new photo to DB
func PhotosInsert(path string, id string) error {
	stmt, err := db.Prepare("INSERT INTO photos(path, id) VALUES(?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path, id)
	if err != nil {
		return fmt.Errorf("Can't insert photo path:%s error:%s", path, err)
	}
	return nil
}

// PhotosDelete deletes a photo from DB
func PhotosDelete(id string) error {
	stmt, err := db.Prepare("DELETE FROM photos WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return  fmt.Errorf("Can't delete photo. id=%s", id)
	}
	return nil
}

// PhotosGetEmptySet returns photos without set
// TODO may be it is better to return slice of struct??
func PhotosGetEmptySet() ([][]string, error) {
	res := [][]string{}

	rows, err := db.Query("SELECT id, path FROM photos WHERE set_id is NULL ORDER BY path")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id, path string
	for rows.Next() {
		err := rows.Scan(&id, &path)
		if err != nil {
			return nil, fmt.Errorf("PhotosGetEmptySet: %s", err)
		}
		//res[id] = path
		res = append(res, []string{id, path})
	}
	return res, nil
}

// PhotosAddToSet adds photo to set
func PhotosAddToSet(id, setID string) error {
	stmt, err := db.Prepare("UPDATE photos SET set_id=? WHERE id=?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(setID, id)

	if err != nil {
		return fmt.Errorf("Can't add photo %s to set %s", id, setID)
	}
	return nil
}