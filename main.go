package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/cors"
)

func main() {
	S := Server{
		MyDrive: DriveService{},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "Welcome to the Go Drive API\n")
		if r.URL.Query().Get("id") != "" {
			http.Redirect(w, r, "/reset_token?code="+r.URL.Query().Get("id"), http.StatusTemporaryRedirect)
		}
	})

	mux.HandleFunc("/reset_token", S.ResetToken)
	mux.HandleFunc("/read", ReadFile)
	mux.HandleFunc("/write", S.WriteFile)

	// Setup CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	err := http.ListenAndServe("0.0.0.0:8000", c.Handler(mux))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}
