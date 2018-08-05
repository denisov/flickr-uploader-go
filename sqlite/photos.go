package sqlite

import (
	"log"

	"github.com/pkg/errors"
)

// PhotosInit creates 'photos' table and its indexes
func (s *Service) photosInit() error {
	log.Println("Initing photos table")

	// id has 'text' type (!)
	_, err := s.connection.Exec(`
		CREATE TABLE IF NOT EXISTS photos (
			id text not null primary key,
			path text not null,
			set_id text
		)
	`)
	if err != nil {
		return errors.Wrap(err, "can't create table photos")
	}

	_, err = s.connection.Exec("CREATE UNIQUE INDEX IF NOT EXISTS fileindex ON photos (path)")
	if err != nil {
		return errors.Wrap(err, "can't create index (path) on 'photos' table")
	}

	_, err = s.connection.Exec("CREATE INDEX IF NOT EXISTS setindex ON photos (set_id)")
	if err != nil {
		return errors.Wrap(err, "can't create index (set_id) on 'photos' table")
	}

	return nil
}

// PhotosGetAll returns all photos in DB
func (s *Service) PhotosGetAll() (map[string]string, error) {
	res := map[string]string{}

	rows, err := s.connection.Query("SELECT path, id FROM photos ORDER BY path")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var path string
	var photoID string
	for rows.Next() {
		err := rows.Scan(&path, &photoID)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan row")
		}
		res[path] = photoID
	}
	return res, nil
}

// PhotosInsert inserts new photo to DB
func (s *Service) PhotosInsert(path string, id string) error {
	stmt, err := s.connection.Prepare("INSERT INTO photos(path, id) VALUES(?, ?)")
	// todo ? defer stmt.close ??
	if err != nil {
		return errors.Wrap(err, "Can't prepare")
	}
	_, err = stmt.Exec(path, id)
	if err != nil {
		return errors.Wrapf(err, "Can't insert photo path:%s error:%s", path, err)
	}
	return nil
}

// PhotosDelete deletes a photo from DB
func (s *Service) PhotosDelete(id string) error {
	stmt, err := s.connection.Prepare("DELETE FROM photos WHERE id=?")
	if err != nil {
		return errors.Wrap(err, "Can't prepare")
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrapf(err, "Can't delete photo. id=%s", id)
	}
	return nil
}

// // PhotosGetEmptySet returns photos without set
// // TODO may be it is better to return slice of struct??
// func (s *Service) PhotosGetEmptySet() ([][]string, error) {
// 	res := [][]string{}

// 	rows, err := s.connection.Query("SELECT id, path FROM photos WHERE set_id is NULL ORDER BY path")
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't select photos")
// 	}
// 	defer rows.Close()

// 	var id, path string
// 	for rows.Next() {
// 		err := rows.Scan(&id, &path)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "can't scan row")
// 		}
// 		//res[id] = path
// 		res = append(res, []string{id, path})
// 	}
// 	return res, nil
// }

// PhotosAddToSet add photo to set
func (s *Service) PhotosAddToSet(id, setID string) error {
	stmt, err := s.connection.Prepare("UPDATE photos SET set_id=? WHERE id=?")
	if err != nil {
		return errors.Wrap(err, "can't prepare update")
	}

	_, err = stmt.Exec(setID, id)

	if err != nil {
		return errors.Wrapf(err, "Can't add photo %s to set %s", id, setID)
	}
	return nil
}
