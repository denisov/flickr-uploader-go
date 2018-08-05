package flickruploader

// OauthToken это токен Oauth авторизации
// TODO перенести в пакет flickr
type OauthToken struct {
	Token       string `json:"token"`
	TokenSecret string `json:"token_secret"`
}

type Filemanager interface {
	GetAllPhotos() ([]string, error)
	ParsePath(path string) (relativeDirname, fileName string)
}

type DBStorage interface {
	PhotosGetAll() (map[string]string, error)
	PhotosInsert(path string, id string) error
	PhotosDelete(id string) error
	//PhotosGetEmptySet() ([][]string, error)
	PhotosAddToSet(id, setID string) error
	SetsInsert(id, name string) error
	SetsGetIDByName(name string) (string, error)
}

type RemoteStorage interface {
	UploadPhoto(photoPath string) (string, error)
	DeletePhoto(photoID string) error
	CreatePhotoset(name, photoID string) (string, error)
	AddPhotoToPhotoset(photoID, photosetID string) error
}
