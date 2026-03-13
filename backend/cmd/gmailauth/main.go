package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	credFile := "credentials.json"
	tokenFile := "token.json"

	if len(os.Args) > 1 {
		credFile = os.Args[1]
	}
	if len(os.Args) > 2 {
		tokenFile = os.Args[2]
	}

	credBytes, err := os.ReadFile(credFile)
	if err != nil {
		log.Fatalf("Unable to read %s: %v\nDownload it from Google Cloud Console first.", credFile, err)
	}

	config, err := google.ConfigFromJSON(credBytes, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("==========================================================")
	fmt.Println("Open this URL in your browser and authorize the app:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Print("Paste the authorization code here: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to exchange auth code for token: %v", err)
	}

	f, err := os.Create(tokenFile)
	if err != nil {
		log.Fatalf("Unable to create %s: %v", tokenFile, err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(tok); err != nil {
		log.Fatalf("Unable to write token: %v", err)
	}

	fmt.Printf("\nToken saved to %s\n", tokenFile)
	fmt.Println("You can now start the server — Gmail polling will be enabled.")
}
