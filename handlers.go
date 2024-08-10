package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"google.golang.org/api/drive/v3"
)

type Server struct {
	MyDrive DriveService
}

func (S *Server) ResetToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Reset the token

	switch r.Method {
	case "GET":
	case "POST":
		var payload Payload
		if checkCredentials(r, &payload, w) {
			return
		}
		// save data as credentials.json
		if err := ioutil.WriteFile("credentials.json", *payload.DataBuffer, 0600); err != nil {
			fmt.Fprint(w, "Error Saving Credentials")
			return
		}
	default:
		fmt.Fprint(w, "Method Not Allowed")
	}

	authCode := r.URL.Query().Get("code")
	if authCode != "" {
		if S.MyDrive.authURL != "" {
			if err := S.MyDrive.InitDriveService(); err != nil {
				fmt.Fprint(w, err.Error())
				return
			}
			for {
				if S.MyDrive.authURL != "" {
					break
				}
			}
		}
		http.Redirect(w, r, S.MyDrive.authURL, http.StatusTemporaryRedirect)
	}
	S.MyDrive.authCode = authCode
}

func ReadFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("fileID")
	if len(fileID) == 0 {
		fmt.Fprintf(w, "No File ID Provided\n")
		return
	}
	url := fmt.Sprintf("https://drive.google.com/uc?export=view&id=%s", fileID)

	// Create the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(w, "Unable to retrieve file: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "Failed to download file: HTTP %d", resp.StatusCode)
		return
	}

	// Read the response body
	Data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "Unable to read file: %v", err)
		return
	}
	fmt.Fprint(w, string(Data))
}

type Payload struct {
	HashedPassword string  `json:"password"`
	DataBuffer     *[]byte `json:"data"`
}

func (S *Server) WriteFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("fileID")
	if len(fileID) == 0 {
		fmt.Fprintf(w, "No File ID Provided\n")
		return
	}

	var payload Payload
	if checkCredentials(r, &payload, w) {
		return
	}

	FileName, err := S.MyDrive.GetFileName(fileID)
	if err != nil {
		fmt.Fprintf(w, "Error Getting File Name: %s\n", err)
		return
	}
	file, err := S.MyDrive.service.Files.Update(fileID, &drive.File{
		Name: FileName,
	}).Media(bytes.NewReader(*payload.DataBuffer)).Do()
	if err != nil {
		fmt.Fprintf(w, "Error Writing File ID: %s\n", err)
		return
	}
	fmt.Fprintf(w, "Write File ID: %s\n", file.Id)

}

func checkCredentials(r *http.Request, payload *Payload, w http.ResponseWriter) bool {
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		fmt.Fprintf(w, "Error Reading Request Body: %s\n", err)
		return true
	}
	if payload.HashedPassword != os.Getenv("PASSWORD") {
		fmt.Fprintf(w, "Incorrect Password\n")
		return true
	}
	return false
}
