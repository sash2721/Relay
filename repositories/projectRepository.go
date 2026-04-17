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

func (repo *ProjectRepository) ListProjects(userID string) ([]models.Projects, error) {
	query := `SELECT id, user_id, project_name, repo_url, project_type, COALESCE(active_deployment_id::text, '') AS active_deployment_id, created_at::text, updated_at::text FROM projects WHERE user_id = $1`

	rows, err := repo.DB.Query(context.Background(), query, userID)

	if err != nil {
		return nil, fmt.Errorf("Failed to return the projects list: %w", err)
	}
	defer rows.Close()

	var projects []models.Projects
	for rows.Next() {
		var p models.Projects
		err := rows.Scan(&p.Id, &p.UserId, &p.ProjectName, &p.RepoURL, &p.ProjectType, &p.ActiveDeploymentId, &p.CreatedAt, &p.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("Failed to read the database data: %w", err)
		}

		projects = append(projects, p)
	}

	return projects, nil
}

func (repo *ProjectRepository) GetProject(projectID string) (*models.Projects, error) {
	query := `SELECT id, user_id, project_name, repo_url, project_type, COALESCE(active_deployment_id::text, '') AS active_deployment_id, created_at::text, updated_at::text FROM projects WHERE id = $1`

	row := repo.DB.QueryRow(context.Background(), query, projectID)

	var p models.Projects
	err := row.Scan(&p.Id, &p.UserId, &p.ProjectName, &p.RepoURL, &p.ProjectType, &p.ActiveDeploymentId, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			slog.Debug("Project not found in the DB", slog.String("projectID", projectID))
			return nil, nil
		}

		slog.Error(
			"Error while fetching the project data from the database",
			slog.Any("Error", err),
		)
		return nil, fmt.Errorf("Failed to fetch the project data: %w", err)
	}

	return &p, nil
}

func (repo *ProjectRepository) DeleteProject(projectID string, userID string) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`

	_, err := repo.DB.Exec(context.Background(), query, projectID, userID)

	if err != nil {
		return fmt.Errorf("Failed to delete the project: %w", err)
	}
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

func (repo *ProjectRepository) UpdateActiveDeployment(projectID string, deploymentID string) error {
	query := `UPDATE projects SET active_deployment_id = $1, updated_at = NOW() WHERE id = $2`

	_, err := repo.DB.Exec(context.Background(), query, deploymentID, projectID)

	if err != nil {
		return fmt.Errorf("failed to update active deployment: %w", err)
	}

	return nil
}
