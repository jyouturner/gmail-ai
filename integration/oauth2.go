package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
)

// GetTokenFromJSON loads a token from a file path
func GetTokenFromJSON(path string, config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer tokenFile.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(tokenFile).Decode(token)
	return token, err
}

// SaveTokenToJSON saves a token to a file path
func SaveTokenToJSON(path string, token *oauth2.Token) {
	tokenFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer tokenFile.Close()

	json.NewEncoder(tokenFile).Encode(token)
}

// RequestNewToken requests a new token from the user
func RequestNewToken(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}
