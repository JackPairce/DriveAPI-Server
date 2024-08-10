package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveService struct {
	client  *http.Client
	service *drive.Service

	authURL  string
	authCode string
}

// InitDriveService initializes the DriveService by reading the client secret file, obtaining the client configuration,
// creating the HTTP client, and retrieving the Drive service.
func (D *DriveService) InitDriveService() error {
	ctx := context.Background()

	// Read the client secret file
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	// Parse the client secret file to obtain the client configuration
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	D.client = D.getClient(config)

	// Create the Drive service using the HTTP client
	srv, err := drive.NewService(ctx, option.WithHTTPClient(D.client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Drive client: %v", err)
	}
	D.service = srv
	log.Println("Drive Service Initialized")
	return nil
}

func (D *DriveService) GetFileID(DirectoryName string, fileName string) (string, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents", fileName, DirectoryName)
	fileList, err := D.service.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return "", err
	}
	if len(fileList.Files) == 0 {
		return "", fmt.Errorf("no files found")
	}
	return fileList.Files[0].Id, nil
}

func (D *DriveService) GetFileName(fileID string) (string, error) {
	file, err := D.service.Files.Get(fileID).Fields("name").Do()
	if err != nil {
		return "", err
	}
	return file.Name, nil
}

// getClient retrieves a token, saves the token, then returns the generated client.
func (D *DriveService) getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = D.getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func (D *DriveService) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	D.authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	for {
		if D.authCode != "" {
			break
		}
	}
	//  clean up
	D.authCode = ""
	D.authURL = ""

	tok, err := config.Exchange(context.TODO(), D.authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to create token file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
