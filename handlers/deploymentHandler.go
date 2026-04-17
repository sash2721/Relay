package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sash2721/Relay/services"
)

type DeploymentHandler struct {
	Service *services.DeploymentService
}

func (s *DeploymentHandler) HandleTriggerDeployment(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	responseJson, err, errJson, responseCode := s.Service.TriggerDeployment(userID, projectID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseJson)

	slog.Info("Deployment Triggered Successfully!",
		slog.String("UserID", userID),
		slog.String("ProjectID", projectID),
	)
}

func (s *DeploymentHandler) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	responseJson, err, errJson, responseCode := s.Service.ListDeployments(userID, projectID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseJson)

	slog.Info("Successfully listed the Deployments!",
		slog.String("UserID", userID),
		slog.String("ProjectID", projectID),
	)
}

func (s *DeploymentHandler) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	deploymentID := chi.URLParam(r, "deploymentID")

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	responseJson, err, errJson, responseCode := s.Service.GetDeployment(userID, projectID, deploymentID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseJson)

	slog.Info("Successfully fetched the Deployment!",
		slog.String("UserID", userID),
		slog.String("ProjectID", projectID),
		slog.String("DeploymentID", deploymentID),
	)
}

func (s *DeploymentHandler) HandleDeleteDeployment(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	deploymentID := chi.URLParam(r, "deploymentID")

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	responseJson, err, errJson, responseCode := s.Service.DeleteDeployment(userID, projectID, deploymentID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		w.Write(errJson)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(responseJson)

	slog.Info("Successfully deleted the Deployment!",
		slog.String("UserID", userID),
		slog.String("ProjectID", projectID),
		slog.String("DeploymentID", deploymentID),
	)
}
