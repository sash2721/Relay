package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sash2721/Relay/models"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (repo *AuthRepository) StoreLocalUser(user models.Users) error {
	query := `INSERT INTO users (id, email, country, name, role, password_hash, provider, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := repo.DB.Exec(context.Background(), query, user.Id, user.Email, user.Country, user.Name, user.Role, user.PasswordHash, user.Provider, user.CreatedAt)

	if err != nil {
		return fmt.Errorf("Failed to store the user data in the DB: %w", err)
	}

	return nil
}

func (repo *AuthRepository) StoreOauthUser(user models.Users) error {
	query := `INSERT INTO users (id, email, country, name, role, password_hash, provider, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := repo.DB.Exec(context.Background(), query, user.Id, user.Email, user.Country, user.Name, user.Role, nil, user.Provider, user.CreatedAt)

	if err != nil {
		return fmt.Errorf("Failed to store the user data in the DB: %w", err)
	}

	return nil
}

func (repo *AuthRepository) GetUser(email string) (*models.Users, error) {
	query := `SELECT * FROM users WHERE email = $1`

	row := repo.DB.QueryRow(context.Background(), query, email)

	var user models.Users
	err := row.Scan(&user.Id, &user.Email, &user.Country, &user.Name, &user.Role, &user.PasswordHash, &user.Provider, &user.CreatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			slog.Debug("User not present in the DB for given mail", slog.String("Email:", email))
			return nil, nil
		}

		slog.Error(
			"Error while retrieving the data from the DB",
			slog.Any("Error", err),
		)
		return nil, err
	}

	return &user, nil
}
