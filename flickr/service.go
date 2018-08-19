package flickr

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photos"
	"gopkg.in/masci/flickr.v2/photosets"
)

// Service это сервис для работы с Flickr
type Service struct {
	client          *flickr.FlickrClient
	tokenFile       string
	APIRequestSleep time.Duration
}

// NewService создаёт новый сервис для для работы с flickr
func NewService(APIKey, APISecret, tokenFile string, APIRequestSleep time.Duration) (*Service, error) {

	client := flickr.NewFlickrClient(APIKey, APISecret)

	return &Service{
		client:          client,
		tokenFile:       tokenFile,
		APIRequestSleep: APIRequestSleep,
	}, nil
}

// SetToken загружает токен из файла или запрашивает новый если нет файла
func (s *Service) SetToken() error {
	if s.checkTokenFileExists() {
		log.Printf("Токен файл найден, загружаем данные из него")
		token, err := s.loadToken()
		if err != nil {
			return errors.Wrap(err, "can't load token")
		}

		s.client.OAuthToken = token.Token
		s.client.OAuthTokenSecret = token.TokenSecret
		return nil
	}
	log.Printf("Токен НЕ файл найден, запрашиваем новый токен")

	token, err := s.requestToken()
	if err != nil {
		return errors.Wrap(err, "can't request token")
	}

	log.Printf("Сохраняем токен")
	err = s.saveToken(token)
	if err != nil {
		return errors.Wrap(err, "can't save token")
	}
	return nil

}

// UploadPhoto загружает фото на flickr
func (s *Service) UploadPhoto(photoPath string) (string, error) {
	time.Sleep(s.APIRequestSleep)
	params := flickr.NewUploadParams()
	params.Tags = []string{"flickruploadergo"}

	response, err := flickr.UploadFile(s.client, photoPath, params)
	// иногода flickr 500-тит.
	if response.ErrorCode() == -1 {
		log.Printf(
			"Got -1 error code from Flickr library. It could be an InternalError (500). ResponseErrorMessage: %s. Error: %s Sleep and try again....",
			response.ErrorMsg(),
			err,
		)
		time.Sleep(20 * time.Second)
		response, err = flickr.UploadFile(s.client, photoPath, params)
	}

	if err != nil {
		err = errors.Wrapf(err, "Upload failed. Photo:%s", photoPath)
		if response != nil {
			err = errors.Wrapf(err, "Code:%d Message:%s", response.ErrorCode(), response.ErrorMsg())
		}
		return "", err
	}

	return response.ID, nil
}

// DeletePhoto удаляет фото на flickr
func (s *Service) DeletePhoto(photoID string) error {
	time.Sleep(s.APIRequestSleep)
	_, err := photos.Delete(s.client, photoID)
	if err != nil {
		return errors.Wrap(err, "can't delete photo on flickr")
	}
	return nil
}

// CreatePhotoset создаёт альбом
func (s *Service) CreatePhotoset(name, photoID string) (string, error) {
	time.Sleep(s.APIRequestSleep)
	response, err := photosets.Create(s.client, name, "", photoID)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"Can't create photoset. Name:%s. Photo:%s.",
			name,
			photoID,
		)
	}
	return response.Set.Id, nil
}

// AddPhotoToPhotoset добавляет фото в фотосет
func (s *Service) AddPhotoToPhotoset(photoID, photosetID string) error {
	time.Sleep(s.APIRequestSleep)
	response, err := photosets.AddPhoto(s.client, photosetID, photoID)
	if err != nil {
		if response != nil && response.ErrorCode() == 3 {
			log.Printf(
				"Photo already in set on Flickr. Set:%s. Photo:%s.",
				photosetID,
				photoID,
			)
		} else {
			return errors.Wrapf(
				err,
				"Failed to add photo to set. Set:%s. Photo:%s. Error:%s",
				photosetID,
				photoID,
			)
		}
	}
	return nil
}
