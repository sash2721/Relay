package configs

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOAuthConfig *oauth2.Config
	GithubOAuthConfig *oauth2.Config
)

func InitProviders() {
	sc := GetServerConfig()

	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     sc.GoogleClientID,
		ClientSecret: sc.GoogleClientSecret,
		RedirectURL:  sc.AppURL + sc.GoogleCallbackAPI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		Endpoint: google.Endpoint,
	}

	GithubOAuthConfig = &oauth2.Config{
		ClientID:     sc.GithubClientID,
		ClientSecret: sc.GithubClientSecret,
		RedirectURL:  sc.AppURL + sc.GithubCallbackAPI,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}
}
