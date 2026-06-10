package auth

import (
	"fmt"
	"megome/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.Envs.GoogleOauthClientId,
		ClientSecret: config.Envs.GoogleOauthSecret,
		RedirectURL:  fmt.Sprintf("%s/api/v1/auth/google/callback", config.Envs.BackendUrl),
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}
