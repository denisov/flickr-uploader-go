package main

import (
	"log"
	"fmt"
	"bufio"
	"os"
	"gopkg.in/masci/flickr.v2"
	"encoding/json"
	"io/ioutil"
)

type oauthToken struct {
	Token       string `json:"token"`
	TokenSecret string `json:"token_secret"`
}

func requestToken(client *flickr.FlickrClient) *oauthToken {
	requestTok, err := flickr.GetRequestToken(client)
	if err != nil {
		log.Fatalln(err)
	}
	url, err := flickr.GetAuthorizeUrl(client, requestTok)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Please follow this URL to auth: " + url)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter CODE: ")
	code, _ := reader.ReadString('\n')

	accessTok, err := flickr.GetAccessToken(client, requestTok, code)
	if err != nil {
		log.Fatalln(err)
	}

	return &oauthToken{accessTok.OAuthToken, accessTok.OAuthTokenSecret}
}

func (token *oauthToken) saveToken(fileName string)  {
	jsonString, err := json.MarshalIndent(token, "", "    ")
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(fileName, jsonString, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func loadToken(fileName string) *oauthToken  {
	jsonString, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	token := &oauthToken{}
	err = json.Unmarshal(jsonString, token)
	if err != nil {
		log.Fatalln(err)
	}

	return token
}


