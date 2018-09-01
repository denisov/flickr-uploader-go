# Flickr uploader

Upload a directory of photos to Flickr to use as a backup to your local storage.


* Uploads only `jpg` images to Flickr
* Ignores unwanted directories
* Creates "Sets" (Albums) based on folder name the photo is in



## Setup:
* Create the config from template `$ cp config.yml.dist config.yml`
* Go to <https://www.flickr.com/services/apps> and create new app. Fill `api_key` and `api_secret` in `config.yml`
* Make sure `token_file_name` and `db_path` paths are writable
* Set `photos_path` dir
* Build binary with `go install` 

## Usage:
* run binary `flickr-uploader-go -config /path/to/config.yml`
* default config `config.yml` in current directory

## SystemD setup:
    mkdir -p ~/.config/systemd/user/
    cp systemd-units/* ~/.config/systemd/user/
    systemctl --user enable flickr-uploader-go.timer
    systemctl --user start flickr-uploader-go.timer
    systemctl --user status flickr-uploader-go.timer

Inspired by <https://github.com/trickortweak/flickr-uploader>