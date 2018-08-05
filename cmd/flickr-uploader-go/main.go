package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/denisov/flickr-uploader-go/flickr"
	"github.com/denisov/flickr-uploader-go/photofiles"
	"github.com/denisov/flickr-uploader-go/sqlite"
	"github.com/denisov/flickr-uploader-go/uploader"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configFile := flag.String("config", "config.yml", "path to config.yml")
	flag.Parse()

	config, err := newConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	sqliteService, err := sqlite.NewService(config.DbPath)
	if err != nil {
		log.Fatalf("Can't create sqlite service: %+v", err)
	}

	photofilesService := photofiles.NewService(config.PhotosPath, config.ExcludeDirs)

	flickrService, err := flickr.NewService(
		config.APIKey,
		config.APISecret,
		config.TokenFileName,
		time.Duration(config.APIRequestSleepMs)*time.Millisecond,
	)
	if err != nil {
		log.Fatalf("Can't create flickr service %+v", err)
	}
	err = flickrService.SetToken()
	if err != nil {
		log.Fatalf("Can't set flickr token %+v", err)
	}

	uploader := uploader.NewService(
		photofilesService,
		sqliteService,
		flickrService,
	)
	err = uploader.InitPhotos()
	if err != nil {
		log.Fatal(err)
	}

	uploader.SetFilesToProcess()

	err = uploader.Upload()
	if err != nil {
		log.Fatal(err)
	}

	err = uploader.Delete()
	if err != nil {
		log.Fatal(err)
	}
}
