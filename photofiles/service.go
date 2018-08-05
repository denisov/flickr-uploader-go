package photofiles

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/denisov/flickr-uploader-go"
	"github.com/pkg/errors"
)

type Service struct {
	path        string
	excludeDirs []string
}

// NewService создаёт сервис для доступа к фотофайлам
func NewService(path string, excludeDirs []string) *Service {
	return &Service{
		path:        path,
		excludeDirs: excludeDirs,
	}
}

// GetAllPhotos возвращает все фотографии с путями по алфавиту
func (s *Service) GetAllPhotos() ([]string, error) {
	var photos []string

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "Can't read dir: %s", s.path)
	}

	visit := func(path string, f os.FileInfo, err error) error {
		if flickruploader.StringInSlice(f.Name(), s.excludeDirs) {
			return filepath.SkipDir
		}

		if f.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) != ".jpg" {
			return nil
		}

		photos = append(photos, path)
		return nil
	}
	err := filepath.Walk(s.path, visit)

	return photos, err
}

// ParsePath парсит путь к файлу относительно базовой директории
// возвращает имя директории относительно базовой и имя файла
func (s *Service) ParsePath(path string) (relativeDirname, fileName string) {
	relPath, err := filepath.Rel(s.path, path)
	if err != nil {
		log.Fatalln(err)
	}
	relativeDirname = filepath.Dir(relPath)
	fileName = filepath.Base(relPath)
	return
}
