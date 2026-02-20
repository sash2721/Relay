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
)

func main() {
	fmt.Println("Relay starts!")
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{ "message": "Relay Service Started" }`))
	})

	// getting the configs
	serverConfig := configs.GetServerConfig()

	// initialising the server
	var server *http.Server

	// creating a dev environment server
	if serverConfig.Env == "development" {
		server = &http.Server{
			Addr:         serverConfig.Port,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}

	// creating a channel to listen for OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() // cancel the context at the end

	// starting the server in a goroutine (asynchronous)
	go func() {
		fmt.Printf("Relay Backend Server listening on PORT%s\n", server.Addr)
		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			slog.Error(
				"Error while starting the Server:",
				slog.Any("Error:", err),
			)
		}
	}()

	// Block here and wait for the OS Background signals
	<-ctx.Done()

	// if any signal comes then log the message for shutting down the server
	slog.Info("Shutdown Signal received, shutting down the backend server gracefully!")

	// creating a context with 5 seconds timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// shutdown using the shutdown context (Attempting graceful shutdown)
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error(
			"Server forced to shutdown:",
			slog.Any("Error", err),
		)
	}

	slog.Info("Server Exited!")
}
