package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sash2721/Relay/configs"
	"github.com/sash2721/Relay/db"
	"github.com/sash2721/Relay/handlers"
	"github.com/sash2721/Relay/middlewares"
	"github.com/sash2721/Relay/proxy"
	"github.com/sash2721/Relay/repositories"
	"github.com/sash2721/Relay/services"
)

func main() {
	slog.Info("Relay Starts!🚀")

	configs.InitServerConfig()
	configs.InitProviders()

	r := chi.NewRouter()

	// common middlewares for all routes here
	r.Use(middlewares.LoggingMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{ "message": "Relay Backend Service Running" }`))
	})

	serverConfig := configs.GetServerConfig()

	// creating db connection and running the migrations
	err := db.Connect(serverConfig.DbConnectionString)
	if err != nil {
		slog.Error("DB connection not established!", slog.Any("Error", err))
		os.Exit(1)
	}
	defer db.Close()

	err = db.RunMigrations()
	if err != nil {
		slog.Error(
			"Failed to run the migrations",
			slog.Any("Error:", err),
		)
		os.Exit(1)
	}

	// initialising the log streamer
	logStreamer := services.NewLogStreamer()

	// start cleanup job
	services.StartCleanupJob()

	// creating repositories
	authRepository := repositories.NewAuthRepository(db.Pool)
	projectRepository := repositories.NewProjectRepository(db.Pool)
	deploymentRepository := repositories.NewDeploymentRepository(db.Pool)

	// creating services
	authService := services.NewAuthService(authRepository)
	projectService := services.NewProjectService(projectRepository)
	builderService := services.NewBuilderService(logStreamer)
	deploymentService := services.NewDeploymentService(deploymentRepository, projectRepository, builderService, logStreamer)

	// creating handlers and injecting services into them
	authHandler := handlers.AuthHandler{Service: authService}
	projectHandler := handlers.ProjectHandler{Service: projectService}
	logStreamHandler := handlers.LogStreamHandler{LogStreamer: logStreamer}
	deploymentHandler := handlers.DeploymentHandler{Service: deploymentService}

	// public routes
	r.Post(serverConfig.LoginAPI, authHandler.HandleLogin)
	r.Post(serverConfig.SignupAPI, authHandler.HandleSignup)
	r.Get(serverConfig.GoogleLoginAPI, authHandler.HandleGoogleLogin)
	r.Get(serverConfig.GoogleCallbackAPI, authHandler.HandleGoogleCallback)
	r.Get(serverConfig.GithubLoginAPI, authHandler.HandleGithubLogin)
	r.Get(serverConfig.GithubCallbackAPI, authHandler.HandleGithubCallback)
	r.Get(serverConfig.LogoutAPI, handlers.HandleLogout)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthZMiddleware)
		r.Use(middlewares.AuthNMiddleware)

		// add protected routes here
		// project handler
		r.Post(serverConfig.ProjectAPI, projectHandler.HandleCreateProject)
		r.Get(serverConfig.ProjectAPI, projectHandler.HandleListProjects)
		r.Get(serverConfig.UpdateProjectAPI, projectHandler.HandleGetProject)
		r.Delete(serverConfig.UpdateProjectAPI, projectHandler.HandleDeleteProject)

		// log streamer
		r.Get(serverConfig.StreamLogsAPI, logStreamHandler.HandlerLogStream)

		// deployment handler
		// adding rate limit to the trigger deployment
		r.Post(serverConfig.TriggerDeploymentAPI,
			middlewares.RateLimitMiddleware(
				http.HandlerFunc(deploymentHandler.HandleTriggerDeployment),
			).ServeHTTP,
		)
		r.Get(serverConfig.GetDeploymentAPI, deploymentHandler.HandleGetDeployment)
		r.Get(serverConfig.ListDeploymentsAPI, deploymentHandler.HandleListDeployments)
		r.Delete(serverConfig.DeleteDeploymentAPI, deploymentHandler.HandleDeleteDeployment)
	})

	var server *http.Server

	if serverConfig.Env == "development" {
		server = &http.Server{
			Addr:         serverConfig.Port,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 0,
			IdleTimeout:  60 * time.Second,
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("Relay Backend Server listening on PORT%s\n", server.Addr)
		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			slog.Error("Error while starting the Server:",
				slog.Any("Error:", err),
			)
		}
	}()

	proxyServer := &http.Server{
		Addr:    serverConfig.ProxyPort,
		Handler: proxy.NewProxyHandler(deploymentRepository),
	}

	go func() {
		slog.Info("Proxy server listening", slog.String("port", serverConfig.ProxyPort))
		proxyServer.ListenAndServe()
	}()

	<-ctx.Done()

	slog.Info("Shutdown Signal received, shutting down the backend server gracefully!")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("Server forced to shutdown:",
			slog.Any("Error", err),
		)
	}

	proxyServer.Shutdown(shutdownCtx)

	slog.Info("Server Exited!")
}
