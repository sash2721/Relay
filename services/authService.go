package services

import (
	"log/slog"
	"net/http"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/utils"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(email string, password string) (string, string, string, string, string, error, []byte, int) {
	slog.Debug("Login called", slog.String("Email", email))

	// TODO: fetch user from repository by email
	// TODO: compare password using utils.ComparePassword(hashedPassword, password)

	// TODO: get actual userID, userName, role from repository
	var userID string
	var userName string
	var role string

	jwtToken, err := utils.GenerateToken(userID, email, role)
	if err != nil {
		slog.Error("Error while generating the JWT token", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while generating the JWT token", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	return jwtToken, userID, userName, email, role, nil, nil, 0
}

func (s *AuthService) Signup(name string, email string, password string, country string) (string, string, string, string, string, error, []byte, int) {
	slog.Debug("Signup called", slog.String("Email", email))

	// TODO: check if user already exists in repository

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("Error while hashing the password", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while hashing the password", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	_ = hashedPassword
	// TODO: store user in repository with hashedPassword, name, email, country

	// Auto-login after signup
	jwtToken, userID, userName, userEmail, role, loginErr, errJson, errorCode := s.Login(email, password)
	if loginErr != nil {
		return "", "", "", "", "", loginErr, errJson, errorCode
	}

	slog.Debug("JWT Token generated successfully", slog.String("Email", email))
	return jwtToken, userID, userName, userEmail, role, nil, nil, 0
}

func (s *AuthService) OAuthLogin(email string, name string, provider string) (string, string, string, string, string, error, []byte, int) {
	slog.Debug("OAuthLogin called",
		slog.String("Email", email),
		slog.String("Provider", provider),
	)

	// TODO: upsert user in repository (create if not exists, update if exists)

	// TODO: get actual userID, role from repository
	var userID string
	var role string

	jwtToken, err := utils.GenerateToken(userID, email, role)
	if err != nil {
		slog.Error("Error while generating the JWT token", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while generating the JWT token", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	return jwtToken, userID, name, email, role, nil, nil, 0
}

func (s *AuthService) ValidateLoginRequest(email string, password string) ([]byte, int) {
	if email == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Email is required", nil)
		return errorJson, badRequestError.Code
	}
	if password == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Password is required", nil)
		return errorJson, badRequestError.Code
	}
	return nil, 0
}

func (s *AuthService) ValidateSignupRequest(name string, email string, password string, country string) ([]byte, int) {
	if name == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Name is required", nil)
		return errorJson, badRequestError.Code
	}
	if email == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Email is required", nil)
		return errorJson, badRequestError.Code
	}
	if password == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Password is required", nil)
		return errorJson, badRequestError.Code
	}
	if country == "" {
		errorJson, badRequestError := errors.NewBadRequestError("Country is required", nil)
		return errorJson, badRequestError.Code
	}
	return nil, 0
}

func (s *AuthService) LoginResponse(token string, userID string, name string, email string, role string) map[string]any {
	return map[string]any{
		"token":  token,
		"userID": userID,
		"name":   name,
		"email":  email,
		"role":   role,
	}
}

func (s *AuthService) SignupResponse(token string, userID string, name string, email string, role string) map[string]any {
	return map[string]any{
		"token":  token,
		"userID": userID,
		"name":   name,
		"email":  email,
		"role":   role,
	}
}

func (s *AuthService) ErrorResponse(errJson []byte, errorCode int) ([]byte, int) {
	if errorCode == http.StatusBadRequest {
		return errJson, http.StatusBadRequest
	}
	return errJson, http.StatusInternalServerError
}
