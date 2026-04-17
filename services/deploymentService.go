package services

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/models"
	"github.com/sash2721/Relay/repositories"
)

type DeploymentService struct {
	DepRepo     *repositories.DeploymentRepository
	ProjRepo    *repositories.ProjectRepository
	BodyBuilder *BuilderService
	YTStreamer  *LogStreamer
}

func NewDeploymentService(depRepo *repositories.DeploymentRepository, projRepo *repositories.ProjectRepository, builder *BuilderService, streamer *LogStreamer) *DeploymentService {
	return &DeploymentService{
		DepRepo:     depRepo,
		ProjRepo:    projRepo,
		BodyBuilder: builder,
		YTStreamer:  streamer,
	}
}

func (s *DeploymentService) TriggerDeployment(userID string, projectID string) ([]byte, error, []byte, int) {
	projDetails, ownerErr, errJsonData, errCode := s.verifyProjectOwnership(userID, projectID)
	if ownerErr != nil {
		return nil, ownerErr, errJsonData, errCode
	}

	// create a deployment record with pending status
	createDeployDetails := models.Deployments{
		ProjectId: projectID,
		Status:    "pending",
	}

	deploymentDetails, err := s.DepRepo.CreateDeployment(createDeployDetails)

	if err != nil {
		slog.Error(
			"Failed to create deployment",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to create deployment", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// initiate goroutine for deployment pipeline
	go func() {
		deploymentID := deploymentDetails.Id

		// clone
		s.DepRepo.UpdateDeploymentStatus(deploymentID, "cloning")
		clonePath, cloneErr, _, _ := s.BodyBuilder.Clone(projDetails.RepoURL)
		if cloneErr != nil {
			slog.Error("Clone failed", slog.Any("Error", cloneErr))
			s.DepRepo.UpdateDeploymentFailed(deploymentID, "Failed to clone repository: "+cloneErr.Error())
			return
		}
		defer os.RemoveAll(clonePath)

		// detect
		s.DepRepo.UpdateDeploymentStatus(deploymentID, "detecting")
		projectType, detectErr, _, _ := s.BodyBuilder.DetectProjectType(clonePath)
		if detectErr != nil {
			slog.Error("Detection failed", slog.Any("Error", detectErr))
			s.DepRepo.UpdateDeploymentFailed(deploymentID, "Failed to detect project type: "+detectErr.Error())
			return
		}

		// build
		s.DepRepo.UpdateDeploymentStatus(deploymentID, "building")
		_, buildErr, _, _ := s.BodyBuilder.Build(clonePath, projectType, deploymentID)
		if buildErr != nil {
			slog.Error("Build failed", slog.Any("Error", buildErr))
			s.DepRepo.UpdateDeploymentFailed(deploymentID, "Build failed: "+buildErr.Error())
			return
		}

		// store
		s.DepRepo.UpdateDeploymentStatus(deploymentID, "deploying")
		buildConfig := s.BodyBuilder.GetBuildConfig(projectType)
		if !buildConfig.IsBackend {
			_, storeErr, _, _ := Store(deploymentID, clonePath, buildConfig.OutputDir)
			if storeErr != nil {
				slog.Error("Storage failed", slog.Any("Error", storeErr))
				s.DepRepo.UpdateDeploymentFailed(deploymentID, "Failed to store artifacts: "+storeErr.Error())
				return
			}
		}

		// mark live
		subdomain := generateSubdomain(projDetails.ProjectName, deploymentID)
		deployedURL := subdomain + ".relay.host"
		s.DepRepo.UpdateDeploymentLive(deploymentID, deployedURL, subdomain)
		s.ProjRepo.UpdateActiveDeployment(projectID, deploymentID)

		slog.Info("Deployment completed successfully",
			slog.String("DeploymentID", deploymentID),
			slog.String("URL", deployedURL),
		)
	}()

	// return deployment ID immediately
	response := models.DeploymentResponse{
		Id:        deploymentDetails.Id,
		ProjectId: deploymentDetails.ProjectId,
		Status:    deploymentDetails.Status,
		CreatedAt: deploymentDetails.CreatedAt,
		UpdatedAt: deploymentDetails.UpdatedAt,
	}

	responseJson, serializeErr, errJson, responseCode := serializeResponse(response, http.StatusCreated)
	if serializeErr != nil {
		return nil, serializeErr, errJson, responseCode
	}

	return responseJson, nil, nil, responseCode
}

func (s *DeploymentService) ListDeployments(userID string, projectID string) ([]byte, error, []byte, int) {
	_, ownerErr, errJsonData, errCode := s.verifyProjectOwnership(userID, projectID)
	if ownerErr != nil {
		return nil, ownerErr, errJsonData, errCode
	}

	// calling the repository for getting details
	deploymentsList, err := s.DepRepo.ListDeployments(projectID)

	if err != nil {
		slog.Error(
			"Failed to fetch deployment list details",
			slog.String("ProjectID:", projectID),
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to fetch deployment list details", err)

		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// structure the data according to the responses
	var deploymentListData []models.DeploymentResponse

	for _, deployment := range deploymentsList {
		deploymentListData = append(deploymentListData, models.DeploymentResponse{
			Id:            deployment.Id,
			ProjectId:     deployment.ProjectId,
			Status:        deployment.Status,
			DeployedURL:   deployment.DeployedURL,
			Subdomain:     deployment.Subdomain,
			FailureReason: deployment.FailureReason,
			CreatedAt:     deployment.CreatedAt,
			UpdatedAt:     deployment.UpdatedAt,
		})
	}

	response := models.DeploymentListResponse{
		Deployments: deploymentListData,
		Count:       len(deploymentListData),
	}

	// serialize the response
	responseData, err, errJsonData, responseCode := serializeResponse(response, http.StatusOK)

	if err != nil {
		return nil, err, errJsonData, responseCode
	}

	return responseData, nil, nil, http.StatusOK
}

func (s *DeploymentService) GetDeployment(userID string, projectID string, deploymentID string) ([]byte, error, []byte, int) {
	_, ownerErr, errJsonData, errCode := s.verifyProjectOwnership(userID, projectID)
	if ownerErr != nil {
		return nil, ownerErr, errJsonData, errCode
	}

	// call the repo for getting the deployment details
	deploymentDetails, err := s.DepRepo.GetDeployment(deploymentID)

	if err != nil {
		slog.Error(
			"Failed to fetch deployment details",
			slog.String("DeploymentID:", deploymentID),
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to fetch deployment details", err)

		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	response := models.DeploymentResponse{
		Id:            deploymentDetails.Id,
		ProjectId:     deploymentDetails.ProjectId,
		Status:        deploymentDetails.Status,
		DeployedURL:   deploymentDetails.DeployedURL,
		Subdomain:     deploymentDetails.Subdomain,
		FailureReason: deploymentDetails.FailureReason,
		CreatedAt:     deploymentDetails.CreatedAt,
		UpdatedAt:     deploymentDetails.UpdatedAt,
	}

	// serialize the response
	responseData, err, errJsonData, responseCode := serializeResponse(response, http.StatusOK)

	if err != nil {
		return nil, err, errJsonData, responseCode
	}

	return responseData, nil, nil, responseCode
}

func (s *DeploymentService) DeleteDeployment(userID string, projectID string, deploymentID string) ([]byte, error, []byte, int) {
	_, ownerErr, errJsonData, errCode := s.verifyProjectOwnership(userID, projectID)
	if ownerErr != nil {
		return nil, ownerErr, errJsonData, errCode
	}

	// call the repository for delete the deployment
	deleteErr := s.DepRepo.DeleteDeployment(deploymentID)

	if deleteErr != nil {
		slog.Error(
			"Failed to delete deployment details",
			slog.String("ProjectID:", projectID),
			slog.Any("Error:", deleteErr),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to delete deployment details", deleteErr)

		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	return nil, nil, nil, http.StatusOK
}

// helper functions
func (s *DeploymentService) verifyProjectOwnership(userID string, projectID string) (*models.Projects, error, []byte, int) {
	projDetails, err := s.ProjRepo.GetProject(projectID)

	if err != nil {
		slog.Error(
			"Failed to fetch project details",
			slog.String("ProjectID:", projectID),
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to fetch project details", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	if projDetails == nil || projDetails.UserId != userID {
		slog.Error(
			"Project not found in the database",
			slog.String("ProjectID:", projectID),
		)
		errJsonData, notFoundError := errors.NewNotFoundError("Project not found", nil)
		return nil, notFoundError, errJsonData, notFoundError.Code
	}

	return projDetails, nil, nil, 0
}

func generateSubdomain(projectName string, deploymentID string) string {
	slug := strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
	shortID := deploymentID[:8]
	return slug + "-" + shortID
}
