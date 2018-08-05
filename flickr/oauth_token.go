package flickr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	flickruploader "github.com/denisov/flickr-uploader-go"
	"github.com/pkg/errors"
	masciflickr "gopkg.in/masci/flickr.v2"
)

// requestToken формирует URL и запрашивает у пользователя авторизацию
func (s *Service) requestToken() (flickruploader.OauthToken, error) {
	requestTok, err := masciflickr.GetRequestToken(s.client)
	if err != nil {
		return flickruploader.OauthToken{}, errors.Wrap(err, "can't get request token")
	}
	url, err := masciflickr.GetAuthorizeUrl(s.client, requestTok)
	if err != nil {
		return flickruploader.OauthToken{}, errors.Wrap(err, "can't get AuthorizeUrl")
	}
	fmt.Println("Please follow this URL to auth: " + url)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter CODE: ")
	code, _ := reader.ReadString('\n')

	accessTok, err := masciflickr.GetAccessToken(s.client, requestTok, code)
	if err != nil {
		return flickruploader.OauthToken{}, errors.Wrap(err, "can't get AccessToken")
	}

	token := flickruploader.OauthToken{
		Token:       accessTok.OAuthToken,
		TokenSecret: accessTok.OAuthTokenSecret,
	}

	return token, nil
}

func (s *Service) checkTokenFileExists() (exists bool) {
	exists = true

	if _, err := os.Stat(s.tokenFile); os.IsNotExist(err) {
		exists = false
	}

	return
}

func (s *Service) saveToken(token flickruploader.OauthToken) error {
	jsonString, err := json.MarshalIndent(token, "", "    ")
	if err != nil {
		return errors.Wrap(err, "can't marshal token file")
	}

	err = ioutil.WriteFile(s.tokenFile, jsonString, 0644)
	if err != nil {
		return errors.Wrap(err, "can't write token file")
	}

	return nil
}

func (s *Service) loadToken() (flickruploader.OauthToken, error) {
	jsonString, err := ioutil.ReadFile(s.tokenFile)
	if err != nil {
		return flickruploader.OauthToken{}, errors.Wrap(err, "can't read file")
	}
	token := &flickruploader.OauthToken{}
	err = json.Unmarshal(jsonString, token)
	if err != nil {
		return flickruploader.OauthToken{}, errors.Wrap(err, "can't unmarshal token file")
	}

	return *token, nil
}
