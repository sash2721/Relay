package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sash2721/Relay/models"
)

type DeploymentRepository struct {
	DB *pgxpool.Pool
}

func NewDeploymentRepository(db *pgxpool.Pool) *DeploymentRepository {
	return &DeploymentRepository{DB: db}
}

func (repo *DeploymentRepository) CreateDeployment(deployment models.Deployments) (*models.Deployments, error) {
	query := `
		INSERT INTO deployments (project_id, status)
		VALUES ($1, $2)
		RETURNING COALESCE(id::text, '') AS id, project_id, status, COALESCE(deployed_url::text, '') AS deployed_url, COALESCE(subdomain::text, '') AS subdomain, COALESCE(failure_reason::text, '') AS failure_reason, created_at::text, updated_at::text
	`

	row := repo.DB.QueryRow(context.Background(), query, deployment.ProjectId, deployment.Status)

	var deploymentDetails models.Deployments
	err := row.Scan(&deploymentDetails.Id, &deploymentDetails.ProjectId, &deploymentDetails.Status, &deploymentDetails.DeployedURL, &deploymentDetails.Subdomain, &deploymentDetails.FailureReason, &deploymentDetails.CreatedAt, &deploymentDetails.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store deployment details: %w", err)
	}

	return &deploymentDetails, nil
}

func (repo *DeploymentRepository) GetDeployment(deploymentID string) (*models.Deployments, error) {
	query := `
		SELECT id, project_id, status, COALESCE(deployed_url, '') AS deployed_url, COALESCE(subdomain, '') AS subdomain, COALESCE(failure_reason, '') AS failure_reason, created_at::text, updated_at::text FROM deployments WHERE id = $1
	`

	row := repo.DB.QueryRow(context.Background(), query, deploymentID)

	var deploymentDetails models.Deployments
	err := row.Scan(&deploymentDetails.Id, &deploymentDetails.ProjectId, &deploymentDetails.Status, &deploymentDetails.DeployedURL, &deploymentDetails.Subdomain, &deploymentDetails.FailureReason, &deploymentDetails.CreatedAt, &deploymentDetails.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment details: %w", err)
	}

	return &deploymentDetails, nil
}

func (repo *DeploymentRepository) ListDeployments(projectID string) ([]models.Deployments, error) {
	query := `
		SELECT id, project_id, status, COALESCE(deployed_url, '') AS deployed_url, COALESCE(subdomain, '') AS subdomain, COALESCE(failure_reason, '') AS failure_reason, created_at::text, updated_at::text FROM deployments WHERE project_id = $1
	`

	rows, err := repo.DB.Query(context.Background(), query, projectID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch all the deployments: %w", err)
	}
	defer rows.Close()

	var deploymentsList []models.Deployments

	for rows.Next() {
		var d models.Deployments

		err := rows.Scan(&d.Id, &d.ProjectId, &d.Status, &d.DeployedURL, &d.Subdomain, &d.FailureReason, &d.CreatedAt, &d.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch all the deployments: %w", err)
		}

		deploymentsList = append(deploymentsList, d)
	}
	return deploymentsList, nil
}

func (repo *DeploymentRepository) UpdateDeploymentStatus(deploymentID string, status string) error {
	query := `
		UPDATE deployments 
		SET status = $1
		WHERE id = $2
	`

	_, err := repo.DB.Exec(context.Background(), query, status, deploymentID)

	if err != nil {
		return fmt.Errorf("failed to update the deployment status: %w", err)
	}

	return nil
}

func (repo *DeploymentRepository) UpdateDeploymentFailed(deploymentID string, reason string) error {
	query := `
		UPDATE deployments
		SET status = 'failed', failure_reason = $1
		WHERE id = $2
	`

	_, err := repo.DB.Exec(context.Background(), query, reason, deploymentID)

	if err != nil {
		return fmt.Errorf("failed to update the deployment failed status & reason: %w", err)
	}

	return nil
}

func (repo *DeploymentRepository) UpdateDeploymentLive(deploymentID string, deployedURL string, subdomain string) error {
	query := `
		UPDATE deployments
		SET status = 'live', deployed_url = $1, subdomain = $2
		WHERE id = $3
	`

	_, err := repo.DB.Exec(context.Background(), query, deployedURL, subdomain, deploymentID)

	if err != nil {
		return fmt.Errorf("failed to update the deployment status: %w", err)
	}

	return nil
}

func (repo *DeploymentRepository) DeleteDeployment(deploymentID string) error {
	query := `
		DELETE FROM deployments
		WHERE id = $1
	`

	_, err := repo.DB.Exec(context.Background(), query, deploymentID)

	if err != nil {
		return fmt.Errorf("failed to delete the deployment: %w", err)
	}

	return nil
}

func (repo *DeploymentRepository) CreateDeploymentLog(log models.DeploymentLogs) error {
	query := `
		INSERT INTO deployment_logs (deployment_id, message)
		VALUES ($1, $2)
	`

	_, err := repo.DB.Exec(context.Background(), query, log.DeploymentId, log.Message)

	if err != nil {
		return fmt.Errorf("failed to store the log details for deployment: %w", err)
	}

	return nil
}

func (repo *DeploymentRepository) GetDeploymentBySubdomain(subdomain string) (*models.Deployments, error) {
	query := `
		SELECT id, project_id, status, COALESCE(deployed_url, '') AS deployed_url, COALESCE(subdomain, '') AS subdomain, COALESCE(failure_reason, '') AS failure_reason, created_at::text, updated_at::text FROM deployments 
		WHERE subdomain = $1 AND status = 'live'
	`

	row := repo.DB.QueryRow(context.Background(), query, subdomain)

	var deploymentDetails models.Deployments
	err := row.Scan(&deploymentDetails.Id, &deploymentDetails.ProjectId, &deploymentDetails.Status, &deploymentDetails.DeployedURL, &deploymentDetails.Subdomain, &deploymentDetails.FailureReason, &deploymentDetails.CreatedAt, &deploymentDetails.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment details: %w", err)
	}

	return &deploymentDetails, nil
}
