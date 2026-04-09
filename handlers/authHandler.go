package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/sash2721/Relay/configs"
	"github.com/sash2721/Relay/services"
)

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

const (
	stateCookieName = "oauth_state"
	jwtCookieName   = "auth_token"
	SessionDuration = 24 * time.Hour

	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	githubUserInfoURL = "https://api.github.com/user"
	githubEmailURL    = "https://api.github.com/user/emails"
)

type OAuthUserInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GithubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

var authService = services.NewAuthService()

func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		slog.Error("Error while generating the state token",
			slog.Any("error", err),
		)
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func isSecureCookie() bool {
	return configs.GetServerConfig().Env != "development"
}

func setJWTCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionDuration.Seconds()),
	})
}

func clearJWTCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     jwtCookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureCookie(),
		MaxAge:   -1,
	})
}

func setStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureCookie(),
		MaxAge:   int((5 * time.Minute).Seconds()),
	})
}

func validateState(w http.ResponseWriter, r *http.Request) bool {
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		slog.Warn("Missing OAuth state cookie")
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:   stateCookieName,
		Path:   "/",
		MaxAge: -1,
	})

	queryState := r.URL.Query().Get("state")
	if queryState == "" || queryState != stateCookie.Value {
		slog.Warn("OAuth state mismatch",
			slog.String("expected", stateCookie.Value),
			slog.String("received", queryState),
		)
		return false
	}

	return true
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("Error while decoding the login request body",
			slog.Any("Error", err),
		)
		http.Error(w, `{"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	errJson, errorCode := authService.ValidateLoginRequest(req.Email, req.Password)
	if errJson != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	token, userID, name, email, role, loginErr, errJson, errorCode := authService.Login(req.Email, req.Password)
	if loginErr != nil {
		slog.Error("Login failed",
			slog.String("Email", req.Email),
			slog.Any("Error", loginErr),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	setJWTCookie(w, token)

	slog.Info("User logged in successfully",
		slog.String("Email", email),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  token,
		UserID: userID,
		Name:   name,
		Email:  email,
		Role:   role,
	})
}

func HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("Error while decoding the signup request body",
			slog.Any("Error", err),
		)
		http.Error(w, `{"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	errJson, errorCode := authService.ValidateSignupRequest(req.Name, req.Email, req.Password, req.Country)
	if errJson != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	token, userID, name, email, role, signupErr, errJson, errorCode := authService.Signup(req.Name, req.Email, req.Password, req.Country)
	if signupErr != nil {
		slog.Error("Signup failed",
			slog.String("Email", req.Email),
			slog.Any("Error", signupErr),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	setJWTCookie(w, token)

	slog.Info("User signed up successfully",
		slog.String("Email", email),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  token,
		UserID: userID,
		Name:   name,
		Email:  email,
		Role:   role,
	})
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("Failed to generate state for Google login",
			slog.Any("error", err),
		)
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	setStateCookie(w, state)

	url := configs.GoogleOAuthConfig.AuthCodeURL(state)
	slog.Info("Redirecting user to Google OAuth consent screen")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !validateState(w, r) {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Warn("Missing authorization code in Google callback")
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := configs.GoogleOAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange Google auth code for token",
			slog.Any("error", err),
		)
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := configs.GoogleOAuthConfig.Client(r.Context(), token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		slog.Error("Failed to fetch user info from Google",
			slog.Any("error", err),
		)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read Google user info response",
			slog.Any("error", err),
		)
		http.Error(w, "failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo OAuthUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		slog.Error("Failed to parse Google user info",
			slog.Any("error", err),
		)
		http.Error(w, "failed to parse user info", http.StatusInternalServerError)
		return
	}

	slog.Info("Google OAuth login successful",
		slog.String("provider", "google"),
	)

	jwtToken, userID, userName, email, role, oauthErr, errJson, errorCode := authService.OAuthLogin(userInfo.Email, userInfo.Name, "google")
	if oauthErr != nil {
		slog.Error("Failed to process Google OAuth login",
			slog.Any("error", oauthErr),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	setJWTCookie(w, jwtToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  jwtToken,
		UserID: userID,
		Name:   userName,
		Email:  email,
		Role:   role,
	})
}

func HandleGithubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		slog.Error("Failed to generate state for GitHub login",
			slog.Any("error", err),
		)
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	setStateCookie(w, state)

	url := configs.GithubOAuthConfig.AuthCodeURL(state)
	slog.Info("Redirecting user to GitHub OAuth consent screen")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if !validateState(w, r) {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Warn("Missing authorization code in GitHub callback")
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := configs.GithubOAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		slog.Error("Failed to exchange GitHub auth code for token",
			slog.Any("error", err),
		)
		http.Error(w, "failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := configs.GithubOAuthConfig.Client(r.Context(), token)
	resp, err := client.Get(githubUserInfoURL)
	if err != nil {
		slog.Error("Failed to fetch user info from GitHub",
			slog.Any("error", err),
		)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read GitHub user info response",
			slog.Any("error", err),
		)
		http.Error(w, "failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo OAuthUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		slog.Error("Failed to parse GitHub user info",
			slog.Any("error", err),
		)
		http.Error(w, "failed to parse user info", http.StatusInternalServerError)
		return
	}

	if userInfo.Email == "" {
		emailResp, err := client.Get(githubEmailURL)
		if err != nil {
			slog.Error("Failed to fetch email from GitHub emails API",
				slog.Any("error", err),
			)
			http.Error(w, "failed to fetch user email", http.StatusInternalServerError)
			return
		}
		defer emailResp.Body.Close()

		emailBody, err := io.ReadAll(emailResp.Body)
		if err != nil {
			slog.Error("Failed to read GitHub emails response",
				slog.Any("error", err),
			)
			http.Error(w, "failed to read email info", http.StatusInternalServerError)
			return
		}

		var emails []GithubEmail
		if err := json.Unmarshal(emailBody, &emails); err != nil {
			slog.Error("Failed to parse GitHub emails",
				slog.Any("error", err),
			)
			http.Error(w, "failed to parse email info", http.StatusInternalServerError)
			return
		}

		for _, e := range emails {
			if e.Primary {
				userInfo.Email = e.Email
				break
			}
		}
	}

	slog.Info("GitHub OAuth login successful",
		slog.String("provider", "github"),
	)

	jwtToken, userID, userName, email, role, oauthErr, errJson, errorCode := authService.OAuthLogin(userInfo.Email, userInfo.Name, "github")
	if oauthErr != nil {
		slog.Error("Failed to process GitHub OAuth login",
			slog.Any("error", oauthErr),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorCode)
		w.Write(errJson)
		return
	}

	setJWTCookie(w, jwtToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:  jwtToken,
		UserID: userID,
		Name:   userName,
		Email:  email,
		Role:   role,
	})
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	clearJWTCookie(w)

	slog.Info("User logged out successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Logged out successfully",
	})
}
