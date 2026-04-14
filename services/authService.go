package services

import (
	"log/slog"
	"net/http"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/models"
	"github.com/sash2721/Relay/repositories"
	"github.com/sash2721/Relay/utils"
)

type AuthService struct {
	Repo *repositories.AuthRepository
}

func NewAuthService(repo *repositories.AuthRepository) *AuthService {
	return &AuthService{Repo: repo}
}

func (s *AuthService) Login(email string, password string) (string, string, string, string, string, error, []byte, int) {
	slog.Debug("Login called", slog.String("Email", email))

	// fetch user from repository by email
	user, err := s.Repo.GetUser(email)

	if err != nil {
		slog.Error(
			"Error while fetching the user from DB",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while fetching the user from DB", err)
		return "", "", "", "", "", internalServerError, errJsonData, internalServerError.Code
	}

	if user == nil {
		slog.Debug(
			"User is not present in the DB",
			slog.String("Email:", email),
		)
		return "", "", "", "", "", nil, nil, 0
	}

	// compare password using utils.ComparePassword(hashedPassword, password)
	err = utils.ComparePassword(user.PasswordHash, password)

	if err != nil {
		slog.Error(
			"Password is incorrect, please retry",
			slog.Any("Error", err),
		)
		errorJson, badRequestError := errors.NewBadRequestError("Password is incorrect, please retry", err)
		return "", "", "", "", "", badRequestError, errorJson, badRequestError.Code
	}

	// get actual userID, userName, role from repository
	var userID string = user.Id
	var userName string = user.Name
	var role string = user.Role

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

	// check if user already exists in repository
	userData, err := s.Repo.GetUser(email)

	if userData != nil {
		slog.Debug("User already exists, please login", slog.String("Email:", email))
		errJsonData, badRequestError := errors.NewBadRequestError("User already exists, please login", nil)
		return "", "", "", "", "", badRequestError, errJsonData, badRequestError.Code
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("Error while hashing the password", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while hashing the password", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	// get the user role from the db
	var userRole string = "user"
	exists, err := s.Repo.CheckUserAdmin(email)

	if err != nil {
		slog.Error("Error while fetching the user role", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while fetching the user role", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	if exists {
		userRole = "admin"
	}

	// store user in repository with hashedPassword, name, email, country
	var user models.Users
	user = models.Users{
		Id:           "",
		Email:        email,
		Country:      country,
		PasswordHash: hashedPassword,
		Name:         name,
		Role:         userRole,
		Provider:     "local",
		CreatedAt:    "",
	}

	err = s.Repo.StoreLocalUser(user)

	if err != nil {
		slog.Error("Error while storing the user info", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while storing the user info", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

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

	// get the user role from the db
	var userRole string = "user"
	exists, err := s.Repo.CheckUserAdmin(email)

	if err != nil {
		slog.Error("Error while fetching the user role", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while fetching the user role", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	if exists {
		userRole = "admin"
	}

	// upsert user in repository (create if not exists, update if exists)
	var user models.Users
	user = models.Users{
		Id:           "",
		Email:        email,
		Country:      "",
		PasswordHash: "",
		Name:         name,
		Role:         userRole,
		Provider:     provider,
		CreatedAt:    "",
	}

	err = s.Repo.UpsertOauthUser(user)

	if err != nil {
		slog.Error("Error while storing the user info", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while storing the user info", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	// get actual userID, role from repository
	userData, err := s.Repo.GetUser(email)

	var userID string = userData.Id

	jwtToken, err := utils.GenerateToken(userID, email, userRole)
	if err != nil {
		slog.Error("Error while generating the JWT token", slog.Any("Error", err))
		errorJson, internalServerError := errors.NewInternalServerError("Error while generating the JWT token", err)
		return "", "", "", "", "", internalServerError, errorJson, internalServerError.Code
	}

	return jwtToken, userID, name, email, userRole, nil, nil, 0
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
