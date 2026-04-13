package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sash2721/Relay/models"
	"github.com/sash2721/Relay/services"
)

type ProjectHandler struct {
	Service *services.ProjectService
}

func (s *ProjectHandler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error(
			"Error while reading the incoming request",
			slog.Any("Error:", err),
		)
		http.Error(w, `{"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// fetching the userID from the request context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// calling the service with the required values
	responseData, err, errJsonData, errCode := s.Service.CreateNewProject(userID, req.ProjectName, req.RepoURL)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errCode)
		w.Write(errJsonData)

		slog.Error(
			"Create new project failed",
			slog.Any("Error:", err),
			slog.Int("Error Code:", errCode),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseData)

	slog.Info(
		"Project created succesfully",
		slog.String("userID:", userID),
	)
}

func (s *ProjectHandler) HandleListProjects(w http.ResponseWriter, r *http.Request) {

}

func (s *ProjectHandler) HandleGetProject(w http.ResponseWriter, r *http.Request) {

}

func (s *ProjectHandler) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {

}
