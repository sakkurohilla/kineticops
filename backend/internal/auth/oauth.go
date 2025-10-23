package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
	Scopes:       []string{"profile", "email"},
	Endpoint:     google.Endpoint,
}

// Usage: redirect users to GoogleOAuthConfig.AuthCodeURL("state")
