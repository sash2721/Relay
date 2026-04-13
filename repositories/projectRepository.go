package repositories

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sash2721/Relay/models"
)

type ProjectRepository struct {
	DB *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{DB: db}
}

func (repo *ProjectRepository) CreateProject(projectData models.Projects) (*models.Projects, error) {
	query := `INSERT INTO projects (user_id, project_name, repo_url, project_type) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, user_id, project_name, repo_url, project_type, COALESCE(active_deployment_id::text, '') AS active_deployment_id, created_at::text, updated_at::text`

	row := repo.DB.QueryRow(context.Background(), query, projectData.UserId, projectData.ProjectName, projectData.RepoURL, projectData.ProjectType)

	var project models.Projects
	err := row.Scan(&project.Id, &project.UserId, &project.ProjectName, &project.RepoURL, &project.ProjectType, &project.ActiveDeploymentId, &project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store project: %w", err)
	}

	return &project, nil
}

func (repo *ProjectRepository) ListProjects(userMail string) []models.Projects {
	return nil
}

func (repo *ProjectRepository) GetProject(projectID string) models.Projects {
	return models.Projects{}
}

func (repo *ProjectRepository) DeleteProject(projectID string) error {
	return nil
}

func (repo *ProjectRepository) UserLookup(userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var userExists bool
	err := repo.DB.QueryRow(context.Background(), query, userID).Scan(&userExists)

	if err != nil {
		slog.Error(
			"Error while fetching the user from the DB",
			slog.Any("Error:", err),
		)
		return false, err
	}

	return userExists, nil
}

func (repo *ProjectRepository) ProjectLookup(repoUrl string, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE repo_url = $1 AND user_id = $2)`

	var projectExists bool
	err := repo.DB.QueryRow(context.Background(), query, repoUrl, userID).Scan(&projectExists)

	if err != nil {
		slog.Error(
			"Error while fetching the project from the DB",
			slog.Any("Error:", err),
		)
		return false, err
	}

	return projectExists, nil
}
