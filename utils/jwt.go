package utils

import (
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sash2721/Relay/configs"
	"github.com/sash2721/Relay/errors"
)

type Claims struct {
	UserID string
	Email  string
	Role   string
}

func GenerateToken(userID string, email string, role string) (string, error) {
	serverConfig := configs.GetServerConfig()
	secretKeyString := serverConfig.SecretKey
	secretKey := []byte(secretKeyString)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId": userID,
			"email":  email,
			"role":   role,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		},
	)

	slog.Debug("Created token with all the required claims",
		slog.String("UserID", userID),
		slog.String("Email", email),
		slog.String("Role", role),
	)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		slog.Error("Error while signing the token with secret key",
			slog.Any("Error", err),
		)
		_, customError := errors.NewInternalServerError("Error while signing the token with secret key", err)
		return "", customError
	}

	slog.Debug("Successfully signed the token with the secret key",
		slog.String("Role", role),
		slog.String("UserID", userID),
	)

	return tokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error, []byte, int) {
	serverConfig := configs.GetServerConfig()
	secretKey := []byte(serverConfig.SecretKey)

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		slog.Error("Error while parsing the token",
			slog.Any("Error", err),
		)
		errorJson, internalServerError := errors.NewInternalServerError("Error while parsing the token", err)
		return nil, internalServerError, errorJson, internalServerError.Code
	}

	if !parsedToken.Valid {
		slog.Error("Invalid Token passed")
		errorJson, badRequestError := errors.NewBadRequestError("Invalid Token passed", nil)
		return nil, badRequestError, errorJson, badRequestError.Code
	}

	slog.Debug("Successfully parsed the token!")

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("Error while converting the claims")
		errorJson, internalServerError := errors.NewInternalServerError("Error while converting the claims", nil)
		return nil, internalServerError, errorJson, internalServerError.Code
	}

	userID := claims["userId"].(string)
	email := claims["email"].(string)
	role := claims["role"].(string)

	slog.Debug("Extracted the claims",
		slog.String("UserID", userID),
		slog.String("Email", email),
		slog.String("Role", role),
	)

	return &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
	}, nil, nil, 0
}
