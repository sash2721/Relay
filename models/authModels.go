package models

import "time"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Country  string `json:"country"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userID"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type OAuthUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GithubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

const (
	StateCookieName = "oauth_state"
	JwtCookieName   = "auth_token"
	SessionDuration = 24 * time.Hour

	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	GithubUserInfoURL = "https://api.github.com/user"
	GithubEmailURL    = "https://api.github.com/user/emails"
)
