package main

import (
	"log"
	"os"
	"github.com/denisov/flickr-uploader-go/models"
	"fmt"
	"gopkg.in/masci/flickr.v2"
	"gopkg.in/masci/flickr.v2/photos"
	"flag"
	"path/filepath"
	"strings"
	"gopkg.in/masci/flickr.v2/photosets"
	"time"
	"sync"
)

var (
	configFile = flag.String("config", "config.yml", "config file name")
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
	config, err := newConfig(*configFile)
	if err != nil {
		log.Fatalln(err);
	}
	models.InitDb(config.DbPath)

	log.Println("Getting all local photos...")
	localPhotos, err := getAllPhotos(config)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Getting all photos in DB")
	dbPhotos, err := models.PhotosGetAll()
	if err != nil{
		log.Fatalln(err)
	}

	// to upload to Flickr - local photos that not in DB
	paths2Upload := []string{}
	for path := range localPhotos {
		if _,ok := dbPhotos[path]; !ok {
			paths2Upload = append(paths2Upload, path)
		}
	}

	// to delete from Flickr - DB photos that not exists anymore
	photoIds2Delete := []string{}
	for path, photoID := range dbPhotos {
		if _,ok := localPhotos[path]; !ok {
			photoIds2Delete = append(photoIds2Delete, photoID)
		}
	}

	client := flickr.NewFlickrClient(config.APIKey, config.APISecret)
	var flickrToken *oauthToken
	if _, err = os.Stat(config.TokenFileName); err == nil {
		log.Println("Token file is found. Loading token file")
		flickrToken = loadToken(config.TokenFileName)
	} else {
		log.Println("Token file is NOT found. Requesting for new token")
		flickrToken = requestToken(client)
		log.Println("Save token to file")
		flickrToken.saveToken(config.TokenFileName)
	}
	client.OAuthToken = flickrToken.Token
	client.OAuthTokenSecret = flickrToken.TokenSecret

	log.Printf("Uploading new photos. Count:%d ..", len(paths2Upload))
	var wg sync.WaitGroup
	for _,photoItem := range paths2Upload {
		wg.Add(1)
		time.Sleep(time.Duration(config.APIRequestSleepMs) * time.Millisecond)
		go func (photoItem string) {
			defer wg.Done()
			uploadClient := flickr.NewFlickrClient(config.APIKey, config.APISecret)
			uploadClient.OAuthToken = flickrToken.Token
			uploadClient.OAuthTokenSecret = flickrToken.TokenSecret

			flickrPhotoID, err := uploadPhotoToFlickr(uploadClient, photoItem)
			if err != nil {
				log.Println(err)
			}
			log.Printf("%s ==> %s ", photoItem, flickrPhotoID)
		} (photoItem)
	}
	wg.Wait()


	log.Printf("Deleting fotos from Flickr. Count: %d ..", len(photoIds2Delete))
	for _, photoID := range photoIds2Delete {
		log.Printf("Deleting photo: %s", photoID)
		time.Sleep(time.Duration(config.APIRequestSleepMs) * time.Millisecond)
		err = deletePhotoFromFlickr(client, photoID)
		if err != nil {
			log.Printf("Error deleting photo %s: %s", photoID, err)
		}
	}

	log.Println("Check and create photosets")
	photosWithoutSet, err := models.PhotosGetEmptySet()
	if err != nil {
		log.Fatalln(err)
	}
	for _,photoItem := range photosWithoutSet {
		photoID := photoItem[0]
		path :=  photoItem[1]

		// there is no need to use routines here
		// this request id quick and there is a chance for race condition
		time.Sleep(time.Duration(config.APIRequestSleepMs) * time.Millisecond)

		relPath,err := filepath.Rel(config.PhotosPath, path)
		if err != nil {
			log.Fatalln(err)
		}
		setName := filepath.Dir(relPath)
		fileName := filepath.Base(relPath)

		err = ensureSetFlickr(client, photoID, setName, fileName)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("Done")
}

func getAllPhotos(config *config) (map[string]bool, error) {
	photosIndexed := map[string]bool{}

	if _, err := os.Stat(config.PhotosPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Can't read dir: %s", config.PhotosPath)
	}

	visit := func(path string, f os.FileInfo, err error) error {
		if stringInSlice(f.Name(), config.ExcludeDirs) {
			return filepath.SkipDir
		}

		if f.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) != ".jpg" {
			return nil
		}

		photosIndexed[path] = true
		return nil
	}
	err := filepath.Walk(config.PhotosPath, visit)

	return photosIndexed, err
}

func uploadPhotoToFlickr(client *flickr.FlickrClient, photoPath string) (string, error) {
	params := flickr.NewUploadParams()
	params.Tags = []string{"flickruploadergo"}

	// upload to Flickr
	response, err := flickr.UploadFile(client, photoPath, params)
	if err != nil {
		err = fmt.Errorf("Upload failed. Photo:%s Err:%s", photoPath, err)
		if response != nil {
			err = fmt.Errorf("%s %s", err, response.ErrorMsg())
		}
		return "", err
	}

	// save to DB
	err = models.PhotosInsert(photoPath, response.ID)
	if err != nil {
		err = fmt.Errorf(
			"PhotoInsert failed. Photo:%s ID:%s. Err:%s",
			photoPath,
			response.ID,
			err,
		)
		return "", err
	}

	return response.ID, nil
}

func deletePhotoFromFlickr(client *flickr.FlickrClient, photoID string) (error) {
	// delete from Flickr
	_, err := photos.Delete(client, photoID)
	if err != nil {
		return err
	}

	// delete from DB
	err = models.PhotosDelete(photoID)
	if err != nil {
		return err
	}

	return nil
}

func ensureSetFlickr(client *flickr.FlickrClient, flickrPhotoID, setName, fileName string) (error) {
	photosetID, err := models.SetsGetIDByName(setName)
	if err != nil {
		return fmt.Errorf("Can't get photoset by name: %s", err)
	}

	if photosetID == "" {
		log.Printf(
			"Photoset '%s' doesn't exists. Create it. Main photo=%s(%s)\n",
			setName,
			fileName,
			flickrPhotoID,
		)
		response, err := photosets.Create(client, setName, "", flickrPhotoID)
		if err != nil {
			return fmt.Errorf(
				"Photoset creation failed. Set:%s. Photo:%s. Error:%s",
				photosetID,
				flickrPhotoID,
				err,
			)
		}
		photosetID = response.Set.Id

		// сохряняем фотосет в БД
		err = models.SetsInsert(photosetID, setName)
		if err != nil {
			return err
		}
	} else {
		log.Printf(
			"Photoset '%s' exists, id=%s. Add photo %s(%s) to photoset",
			setName,
			photosetID,
			fileName,
			flickrPhotoID,
		)
		response,err := photosets.AddPhoto(client, photosetID, flickrPhotoID)
		if err != nil {
			// return the error in all cases except "Photo already in set"
			if response != nil && response.ErrorCode() == 3 {
				fmt.Printf(
					"Photo already in set on Flickr. Set:%s. Photo:%s.",
					photosetID,
					flickrPhotoID,
				)
			} else {
				return fmt.Errorf(
					"Failed to add photo to set. Set:%s. Photo:%s. Error:%s",
					photosetID,
					flickrPhotoID,
					err,
				)
			}
		}
	}

	err = models.PhotosAddToSet(flickrPhotoID, photosetID)
	if err != nil {
		return err
	}

	return  nil
}

