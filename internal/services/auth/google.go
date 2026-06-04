package auth

import (
	"megome/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig = &oauth2.Config{
	ClientID:     config.Envs.GoogleOauthClientId,
	ClientSecret: config.Envs.GoogleOauthSecret,
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes: []string{
		"openid",
		"email",
		"profile",
	},
	Endpoint: google.Endpoint,
}
