package services

import (
	"encoding/json"
	"log/slog"
	"os"
	"regexp"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/models"
	"github.com/sash2721/Relay/repositories"
)

type ProjectService struct {
	Repo        *repositories.ProjectRepository
	BodyBuilder *BuilderService
}

func NewProjectService(repo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{Repo: repo}
}

func (s *ProjectService) CreateNewProject(userID string, projectName string, repoUrl string) ([]byte, error, []byte, int) {
	// check if the user exists in the DB or not
	err, errJsondata, errCode := s.isUserPresent(userID)

	if err != nil {
		return nil, err, errJsondata, errCode
	}

	// validate the project name
	if projectName == "" {
		slog.Error("Project name is required")
		errJsonData, badRequestError := errors.NewBadRequestError("Project name is required", nil)
		return nil, badRequestError, errJsonData, badRequestError.Code
	}

	// validate the GitHub URL sent if its a correct URL or not
	validUrl := isValidGithubURL(repoUrl)

	if !validUrl {
		slog.Error(
			"Invalid URL sent, please send a valid URL",
			slog.String("Repo URL:", repoUrl),
		)
		errJsonData, badRequestError := errors.NewBadRequestError("Invalid URL sent, please send a valid URL", nil)
		return nil, badRequestError, errJsonData, badRequestError.Code
	}

	// do a DB lookup if the project with this link and user already exists or not
	projectExists, err := s.Repo.ProjectLookup(repoUrl, userID)

	if err != nil {
		slog.Error(
			"Error while performing project lookup",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while performing project lookup", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	if projectExists {
		slog.Error(
			"Project trying to be deployed already exists in the database",
			slog.String("Project Name:", projectName),
			slog.String("Project Repo URL:", repoUrl),
			slog.String("UserID:", userID),
		)
		errJsonData, conflictError := errors.NewConflictError("Project trying to be deployed already exists in the database", nil)
		return nil, conflictError, errJsonData, conflictError.Code
	}

	// service to detect the project type
	// clone the github repo
	clonePath, err, errJsonData, responseCode := s.BodyBuilder.Clone(repoUrl)

	if err != nil {
		return nil, err, errJsonData, responseCode
	}
	defer os.RemoveAll(clonePath)

	// determine the project type
	projectType, err, errJsonData, responseCode := s.BodyBuilder.DetectProjectType(clonePath)

	if err != nil {
		return nil, err, errJsonData, responseCode
	}

	// create the projectData object
	projectData := models.Projects{
		UserId:      userID,
		ProjectName: projectName,
		RepoURL:     repoUrl,
		ProjectType: projectType,
	}

	// store the project info in the DB
	repoData, err := s.Repo.CreateProject(projectData)

	if err != nil {
		slog.Error(
			"Error while storing the data in the database",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while storing the data in the database", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// prepare the response data
	response := models.ProjectResponse{
		Id:                 repoData.Id,
		ProjectName:        repoData.ProjectName,
		ProjectType:        repoData.ProjectType,
		RepoURL:            repoData.RepoURL,
		ActiveDeploymentId: repoData.ActiveDeploymentId,
		CreatedAt:          repoData.CreatedAt,
		UpdatedAt:          repoData.UpdatedAt,
	}

	// serializing the response data
	responseJson, err, errJson, responseCode := serializeResponse(response, 201)

	if err != nil {
		return nil, err, errJson, responseCode
	}

	return responseJson, nil, nil, responseCode
}

func (s *ProjectService) ListAllProjects(userID string) ([]byte, error, []byte, int) {
	// check if the user exists in the DB or not
	err, errJsondata, errCode := s.isUserPresent(userID)

	if err != nil {
		return nil, err, errJsondata, errCode
	}

	// call the repository for the projects list
	projectList, err := s.Repo.ListProjects(userID)

	if err != nil {
		slog.Error(
			"Error while fetching the projects from the database",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while fetching the projects from the database", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// create the response list
	var projectResponses []models.ProjectResponse

	for _, project := range projectList {
		projectResponses = append(projectResponses, models.ProjectResponse{
			Id:                 project.Id,
			ProjectName:        project.ProjectName,
			ProjectType:        project.ProjectType,
			RepoURL:            project.RepoURL,
			ActiveDeploymentId: project.ActiveDeploymentId,
			CreatedAt:          project.CreatedAt,
			UpdatedAt:          project.UpdatedAt,
		})
	}

	response := models.ProjectListResponse{
		Projects: projectResponses,
		Count:    len(projectResponses),
	}

	// serializing the response data
	responseJson, err, errJson, responseCode := serializeResponse(response, 200)

	if err != nil {
		return nil, err, errJson, responseCode
	}

	return responseJson, nil, nil, responseCode
}

func (s *ProjectService) GetProject(projectID string, userID string) ([]byte, error, []byte, int) {
	// check if the user exists in the DB or not
	err, errJsondata, errCode := s.isUserPresent(userID)

	if err != nil {
		return nil, err, errJsondata, errCode
	}

	// calling the repository for the project details
	projectData, err := s.Repo.GetProject(projectID)

	if err != nil {
		slog.Error(
			"Error while fetching the project details from the database",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while fetching the project details from the database", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// checking if projectData exists or not
	// and if the projectData is for the right userID
	if projectData == nil || projectData.UserId != userID {
		slog.Debug(
			"Project not found in the database",
		)
		errJsonData, notFoundError := errors.NewNotFoundError("Project not found in the database", nil)
		return nil, notFoundError, errJsonData, notFoundError.Code
	}

	// create the response structure to return
	projectResponse := models.ProjectResponse{
		Id:                 projectID,
		ProjectName:        projectData.ProjectName,
		ProjectType:        projectData.ProjectType,
		RepoURL:            projectData.RepoURL,
		ActiveDeploymentId: projectData.ActiveDeploymentId,
		CreatedAt:          projectData.CreatedAt,
		UpdatedAt:          projectData.UpdatedAt,
	}

	// serializing the response data
	responseJson, err, errJson, responseCode := serializeResponse(projectResponse, 200)

	if err != nil {
		return nil, err, errJson, responseCode
	}

	return responseJson, nil, nil, responseCode
}

func (s *ProjectService) DeleteProject(projectID string, userID string) ([]byte, error, []byte, int) {
	// check if the user exists in the DB or not
	err, errJsondata, errCode := s.isUserPresent(userID)

	if err != nil {
		return nil, err, errJsondata, errCode
	}

	// calling the repository for first getting the projectDetails
	projectData, err := s.Repo.GetProject(projectID)

	if err != nil {
		slog.Error(
			"Error while fetching the project details from the database",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while fetching the project details from the database", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// checking if projectData exists or not
	// and if the projectData is for the right userID
	if projectData == nil || projectData.UserId != userID {
		slog.Debug(
			"Project not found in the database",
		)
		errJsonData, notFoundError := errors.NewNotFoundError("Project not found in the database", nil)
		return nil, notFoundError, errJsonData, notFoundError.Code
	}

	// calling the repository to delete the project
	err = s.Repo.DeleteProject(projectID, userID)

	if err != nil {
		slog.Error(
			"Error while deleting the project",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while deleting the project", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// preparing the response sturcture
	responseData := models.ProjectDeleteResponse{
		Message: "Project Deleted Successfully",
		Success: true,
	}

	// serializing the response data
	responseJson, err, errJson, responseCode := serializeResponse(responseData, 200)

	if err != nil {
		return nil, err, errJson, responseCode
	}

	return responseJson, nil, nil, responseCode
}

// Helper functions
func isValidGithubURL(url string) bool {
	var githubURLRegex = regexp.MustCompile(`^https://github\.com/[\w.\-]+/[\w.\-]+/?$`)

	return githubURLRegex.MatchString(url)
}

func (s *ProjectService) isUserPresent(userID string) (error, []byte, int) {
	userExists, err := s.Repo.UserLookup(userID)

	if err != nil {
		slog.Error(
			"Error while performing user lookup",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while performing user lookup", err)
		return internalServerError, errJsonData, internalServerError.Code
	}

	if !userExists {
		slog.Error(
			"User not found in the database",
			slog.Any("Error:", err),
		)
		errJsonData, notFoundError := errors.NewNotFoundError("User not found in the database", nil)
		return notFoundError, errJsonData, notFoundError.Code
	}

	return nil, nil, 0
}

func serializeResponse(responseData any, statusCode int) ([]byte, error, []byte, int) {
	responseJson, err := json.Marshal(responseData)

	if err != nil {
		slog.Error(
			"Error while marshaling the response data in service",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while marshaling the response data in service", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// return
	return responseJson, nil, nil, statusCode
}
