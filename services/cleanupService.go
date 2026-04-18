package services

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sash2721/Relay/configs"
)

func StartCleanupJob() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cleanupOrphanedContainers()
			cleanupTempDirs()
		}
	}()
	slog.Info("Cleanup job started (runs every hour)")
}

func cleanupOrphanedContainers() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		slog.Error("Cleanup: failed to create Docker client", slog.Any("Error", err))
		return
	}

	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", "relay-build-"), filters.Arg("status", "exited")),
	})
	if err != nil {
		slog.Error("Cleanup: failed to list containers", slog.Any("Error", err))
		return
	}

	for _, c := range containers {
		dockerClient.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true})
		slog.Debug("Cleanup: removed container", slog.String("ID", c.ID[:12]))
	}

	if len(containers) > 0 {
		slog.Info("Cleanup: removed orphaned containers", slog.Int("count", len(containers)))
	}
}

func cleanupTempDirs() {
	tmpDir := os.TempDir()
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		slog.Error("Cleanup: failed to read temp dir", slog.Any("Error", err))
		return
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "repo-") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			// remove if older than 1 hour
			if time.Since(info.ModTime()) > 1*time.Hour {
				os.RemoveAll(filepath.Join(tmpDir, entry.Name()))
				count++
			}
		}
	}

	if count > 0 {
		slog.Info("Cleanup: removed stale temp directories", slog.Int("count", count))
	}
}

func CleanupArtifacts(deploymentID string) {
	serverConfig := configs.GetServerConfig()
	artifactPath := filepath.Join(serverConfig.ArtifactsDir, deploymentID)
	os.RemoveAll(artifactPath)
	slog.Debug("Cleaned up artifacts", slog.String("DeploymentID", deploymentID))
}
