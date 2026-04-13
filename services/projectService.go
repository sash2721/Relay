package services

import (
	"encoding/json"
	"log/slog"
	"regexp"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/models"
	"github.com/sash2721/Relay/repositories"
)

type ProjectService struct {
	Repo *repositories.ProjectRepository
}

func NewProjectService(repo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{Repo: repo}
}

func (s *ProjectService) CreateNewProject(userID string, projectName string, repoUrl string) ([]byte, error, []byte, int) {
	// check if the user exists in the DB or not
	userExists, err := s.Repo.UserLookup(userID)

	if err != nil {
		slog.Error(
			"Error while performing user lookup",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while performing user lookup", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	if !userExists {
		slog.Error(
			"User not found in the database",
			slog.Any("Error:", err),
		)
		errJsonData, notFoundError := errors.NewNotFoundError("User not found in the database", nil)
		return nil, notFoundError, errJsonData, notFoundError.Code
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

	// TODO: write a service to detect the project type (for now hardcoding)
	projectType := "Node.js"

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

	// serialise the response
	response := models.ProjectResponse{
		Id:                 repoData.Id,
		ProjectName:        repoData.ProjectName,
		ProjectType:        repoData.ProjectType,
		RepoURL:            repoData.RepoURL,
		ActiveDeploymentId: repoData.ActiveDeploymentId,
		CreatedAt:          repoData.CreatedAt,
		UpdatedAt:          repoData.UpdatedAt,
	}

	responseData, err := json.Marshal(response)

	if err != nil {
		slog.Error(
			"Error while marshaling the response data in service",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while marshaling the response data in service", err)
		return nil, internalServerError, errJsonData, internalServerError.Code
	}

	// return
	return responseData, nil, nil, 0
}

func (s *ProjectService) ListAllProjects(userName string, userEmail string) ([]byte, error) {
	return nil, nil
}

func (s *ProjectService) GetProject(projectID string) ([]byte, error) {
	return nil, nil
}

func (s *ProjectService) DeleteProject(projectID string) ([]byte, error) {
	return nil, nil
}

// helper functions
func isValidGithubURL(url string) bool {
	var githubURLRegex = regexp.MustCompile(`^https://github\.com/[\w.\-]+/[\w.\-]+/?$`)

	return githubURLRegex.MatchString(url)
}
