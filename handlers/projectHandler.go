package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	responseData, err, errJsonData, responseCode := s.Service.CreateNewProject(userID, req.ProjectName, req.RepoURL)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJsonData)

		slog.Error(
			"Create new project failed",
			slog.Any("Error:", err),
			slog.Int("Error Code:", responseCode),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseData)

	slog.Info(
		"Project created succesfully",
		slog.String("userID:", userID),
	)
}

func (s *ProjectHandler) HandleListProjects(w http.ResponseWriter, r *http.Request) {
	// fetching the userID from the request context
	userID, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// calling the service with the userID
	responseData, err, errJsonData, responseCode := s.Service.ListAllProjects(userID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJsonData)

		slog.Error(
			"Listing all the projects failed",
			slog.Any("Error:", err),
			slog.Int("Error Code:", responseCode),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseData)

	slog.Info(
		"Successfully listed down all the projects!",
		slog.Any("UserID:", userID),
	)
}

func (s *ProjectHandler) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	// fetching the projectID from the URL params
	projectID := chi.URLParam(r, "projectID")

	// featching the userID from the request context
	userID, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// calling the service with the projectID
	responseData, err, errJsonData, responseCode := s.Service.GetProject(projectID, userID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJsonData)

		slog.Error(
			"Fetching the Project failed",
			slog.Any("Error:", err),
			slog.Int("Error Code:", responseCode),
			slog.String("ProjectID", projectID),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseData)

	slog.Info(
		"Successfully fetched the Project Data for given Project ID",
		slog.String("ProjectID:", projectID),
	)
}

func (s *ProjectHandler) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	// fetching the projectID from the URL params
	projectID := chi.URLParam(r, "projectID")

	// featching the userID from the request context
	userID, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// calling the service for deleting the project
	responseData, err, errJsonData, responseCode := s.Service.DeleteProject(projectID, userID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJsonData)

		slog.Error(
			"Failed to delete the project",
			slog.Any("Error:", err),
			slog.Int("Error Code:", responseCode),
			slog.String("ProjectID", projectID),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseData)

	slog.Info(
		"Successfully Deleted the Project Data for given Project ID",
		slog.String("ProjectID:", projectID),
	)
}
