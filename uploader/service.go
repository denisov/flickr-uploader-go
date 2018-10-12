package uploader

import (
	"log"
	"sort"
	"sync"

	"github.com/denisov/flickr-uploader-go"
	"github.com/pkg/errors"
)

// Service это сервис синхронизации файлов на flickr
type Service struct {
	stopped      bool
	mutexStopped sync.Mutex

	photoFiles []string
	dbFiles    map[string]string // FIXME описать или упростить формат

	pathsToUpload    []string // файлы на загрузку
	photoIDsToDelete []string // ID файлов на удаление

	fileManager   flickruploader.Filemanager
	dbStorage     flickruploader.DBStorage
	remoteStorage flickruploader.RemoteStorage
}

// NewService возвращает сервис синхронизации (загрузки)
func NewService(
	fileManager flickruploader.Filemanager,
	dbstorage flickruploader.DBStorage,
	remoteStorage flickruploader.RemoteStorage,
) *Service {
	return &Service{
		fileManager:   fileManager,
		dbStorage:     dbstorage,
		remoteStorage: remoteStorage,
		stopped:       false,
	}
}

// Stop взводит флаг остановки сервиса
func (s *Service) Stop() {
	s.mutexStopped.Lock()
	s.stopped = true
	s.mutexStopped.Unlock()
}

func (s *Service) isStopped() bool {
	s.mutexStopped.Lock()
	isStopped := s.stopped
	s.mutexStopped.Unlock()

	return isStopped
}

// InitPhotos загружает фото из базы и из файловой системы в поля структуры
func (s *Service) InitPhotos() error {

	log.Println("Getting all local photos...")
	photoFiles, err := s.fileManager.GetAllPhotos()
	if err != nil {
		return errors.Wrap(err, "Can't get all photos from disk")
	}

	// фотографии теоретически уже отсортированы, но при появлении русских букв в именах получаются расхождения
	// с sort.Strings и sort.SearchStrings поэтому сортирнём ещё раз
	sort.Strings(photoFiles)
	s.photoFiles = photoFiles

	log.Println("Getting all photos in DB")
	dbFiles, err := s.dbStorage.PhotosGetAll()
	if err != nil {
		errors.Wrap(err, "can't get all photos from DB")
	}
	s.dbFiles = dbFiles

	return nil
}

// SetFilesToProcess определяет файлы для обработки, загрузка и удаление
func (s *Service) SetFilesToProcess() {
	// TODO вынести в ф-ции загрузки и удаления?

	// to upload to Flickr - local photos that not in DB
	for _, path := range s.photoFiles {
		if _, ok := s.dbFiles[path]; !ok {
			s.pathsToUpload = append(s.pathsToUpload, path)
		}
	}

	// когда одновременно добавляются и удаляются фото не будет ли проблем с определением что удалять?

	// to delete from Flickr - DB photos that doesn't exists anymore
	// SearchStrings возвращает индекс искомой строки в отсортированном массиве строк. Если не найдено, то
	// возвращает индекс в который надо вставить
	for path, photoID := range s.dbFiles {
		if idx := sort.SearchStrings(s.photoFiles, path); idx == len(s.photoFiles) || s.photoFiles[idx] != path {
			s.photoIDsToDelete = append(s.photoIDsToDelete, photoID)
		}
	}
}

// Upload загружает фото в удалённое хранилище
func (s *Service) Upload() error {
	if s.isStopped() {
		return nil
	}
	log.Printf("Uploading new photos. Count:%d ..", len(s.pathsToUpload))

	for _, photoPathItem := range s.pathsToUpload {
		if s.isStopped() {
			return nil
		}

		photoID, err := s.remoteStorage.UploadPhoto(photoPathItem)
		if err != nil {
			return errors.Wrapf(err, "Can't upload photo %q", photoPathItem)
		}
		log.Printf("File Uploaded. %s ==> %s ", photoPathItem, photoID)

		err = s.dbStorage.PhotosInsert(photoPathItem, photoID)
		if err != nil {
			return errors.Wrapf(err, "Can't insert photo to db storage %q %q", photoPathItem, photoID)
		}

		// создаём фотосет или добавляем в существующий
		photosetName, fileName := s.fileManager.ParsePath(photoPathItem)
		photosetID, err := s.dbStorage.SetsGetIDByName(photosetName)
		if err != nil {
			return errors.Wrapf(err, "Can't get set name from db storage by name %q", photosetName)
		}
		if photosetID == "" {
			log.Printf("Photoset '%s' doesn't exists. Create it. Main photo=%s(%s)", photosetName, fileName, photoID)
			photosetID, err := s.remoteStorage.CreatePhotoset(photosetName, photoID)
			if err != nil {
				return errors.Wrapf(err, "Can't create photoset %s %s", photosetName, photoID)
			}
			err = s.dbStorage.SetsInsert(photosetID, photosetName)
			if err != nil {
				return errors.Wrapf(err, "Can't insert photoset %s %s", photosetID, photosetName)
			}
		} else {
			log.Printf("Photoset '%s' exists, id=%s. Add photo %s(%s) to photoset", photosetName, photosetID, fileName, photoID)
			err = s.remoteStorage.AddPhotoToPhotoset(photoID, photosetID)
			if err != nil {
				return errors.Wrapf(err, "Can't add photo %q to photoset %q", photoID, photosetID)
			}
		}

		err = s.dbStorage.PhotosAddToSet(photoID, photosetID)
		if err != nil {
			return errors.Wrapf(err, "Can't set photoset %s for photo %s", photosetID, photoID)
		}
	}

	return nil
}

// Delete удаляет фото из удалённого хранилища
func (s *Service) Delete() error {
	if s.isStopped() {
		return nil
	}
	log.Printf("Deleting photos from Flickr. Count: %d ..", len(s.photoIDsToDelete))

	for _, photoID := range s.photoIDsToDelete {
		if s.isStopped() {
			return nil
		}

		log.Printf("Deleting photo: %s", photoID)

		err := s.remoteStorage.DeletePhoto(photoID)
		if err != nil {
			return errors.Wrapf(err, "Can't delete photo %q from remote storage", photoID)
		}

		err = s.dbStorage.PhotosDelete(photoID)
		if err != nil {
			return errors.Wrapf(err, "Can't delete photo %q from db storage", photoID)
		}
	}

	return nil
}
