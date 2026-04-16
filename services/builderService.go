package services

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/models"
)

type BuilderService struct{}

func NewBuilderService() *BuilderService {
	return &BuilderService{}
}

// returns path of the cloned directory
func Clone(repoURL string) (string, error, []byte, int) {
	// create a temp directory
	dirPath, err := os.MkdirTemp("", "repo-*")

	if err != nil {
		slog.Error(
			"Error while creating a temp directory for cloning the repository",
			slog.Any("Error:", err),
		)

		errJsonData, internalServerError := errors.NewInternalServerError("Error while creating a temp directory for cloning the repository", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// prepare the command
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, dirPath)

	// execute the command
	err = cmd.Run()

	if err != nil {
		slog.Error(
			"Error while cloning the repository",
			slog.Any("Error:", err),
		)

		// cleaning up the directory
		os.RemoveAll(dirPath)

		errJsonData, internalServerError := errors.NewInternalServerError("Error while cloning the repository", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// return
	return dirPath, nil, nil, http.StatusCreated
}

func DetectProjectType(cloneDir string) (string, error, []byte, int) {
	// check if the specified file exists in the cloned directory
	// check for Go
	_, err := os.Stat(filepath.Join(cloneDir, "go.mod"))

	if err == nil {
		return "Go", nil, nil, http.StatusOK
	}

	// check for Java (Maven)
	_, err = os.Stat(filepath.Join(cloneDir, "pom.xml"))

	if err == nil {
		return "Java (Maven)", nil, nil, http.StatusOK
	}

	// check for Java (Gradle)
	_, err = os.Stat(filepath.Join(cloneDir, "build.gradle"))

	if err == nil {
		return "Java (Gradle)", nil, nil, http.StatusOK
	}

	// check for Python
	_, err = os.Stat(filepath.Join(cloneDir, "requirements.txt"))

	if err == nil {
		return "Python", nil, nil, http.StatusOK
	}

	_, err = os.Stat(filepath.Join(cloneDir, "pyproject.toml"))

	if err == nil {
		return "Python", nil, nil, http.StatusOK
	}

	// Check for JS / TS related projects
	_, err = os.Stat(filepath.Join(cloneDir, "package.json"))

	if err == nil {
		data, err := os.ReadFile(filepath.Join(cloneDir, "package.json"))
		if err != nil {
			errJsonData, internalServerError := errors.NewInternalServerError("Error while reading package.json", err)
			return "", internalServerError, errJsonData, internalServerError.Code
		}

		var packageData models.PackageJSON
		err = json.Unmarshal(data, &packageData)
		if err != nil {
			errJsonData, internalServerError := errors.NewInternalServerError("Error while parsing package.json", err)
			return "", internalServerError, errJsonData, internalServerError.Code
		}

		// check dependencies for framework detection
		allDeps := make(map[string]string)
		for k, v := range packageData.Dependencies {
			allDeps[k] = v
		}
		for k, v := range packageData.DevDependencies {
			allDeps[k] = v
		}

		if _, ok := allDeps["next"]; ok {
			return "Next.js", nil, nil, http.StatusOK
		}
		if _, ok := allDeps["@angular/core"]; ok {
			return "Angular", nil, nil, http.StatusOK
		}
		if _, ok := allDeps["svelte"]; ok {
			return "Svelte", nil, nil, http.StatusOK
		}
		if _, ok := allDeps["@sveltejs/kit"]; ok {
			return "Svelte", nil, nil, http.StatusOK
		}
		if _, ok := allDeps["vue"]; ok {
			return "Vue", nil, nil, http.StatusOK
		}
		if _, ok := allDeps["react"]; ok {
			return "React", nil, nil, http.StatusOK
		}
		if _, ok := packageData.Scripts["start"]; ok {
			return "Node.js", nil, nil, http.StatusOK
		}
	}

	// check for Static HTML
	_, err = os.Stat(filepath.Join(cloneDir, "index.html"))
	if err == nil {
		return "Static HTML", nil, nil, http.StatusOK
	}

	// fallback
	return "Static HTML", nil, nil, http.StatusOK
}

func GetBuildConfig(projectType string) *models.BuildConfig {
	switch projectType {
	case "Go":
		return &models.BuildConfig{
			BaseImage:  "golang:1.22-alpine",
			InstallCmd: []string{"go", "mod", "download"},
			BuildCmd:   []string{"go", "build", "-o", "app", "."},
			OutputDir:  "",
			IsBackend:  true,
		}
	case "Python":
		return &models.BuildConfig{
			BaseImage:  "python:3.12-slim",
			InstallCmd: []string{"pip", "install", "-r", "requirements.txt"},
			BuildCmd:   []string{"python", "main.py"},
			OutputDir:  "",
			IsBackend:  true,
		}
	case "Java (Maven)":
		return &models.BuildConfig{
			BaseImage:  "eclipse-temurin:21-jdk-alpine",
			InstallCmd: []string{"mvn", "package"},
			BuildCmd:   []string{"sh", "-c", "java -jar target/*.jar"},
			OutputDir:  "",
			IsBackend:  true,
		}
	case "Java (Gradle)":
		return &models.BuildConfig{
			BaseImage:  "eclipse-temurin:21-jdk-alpine",
			InstallCmd: []string{"gradle", "build"},
			BuildCmd:   []string{"sh", "-c", "java -jar build/libs/*.jar"},
			OutputDir:  "",
			IsBackend:  true,
		}
	case "Node.js":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npm", "start"},
			OutputDir:  "",
			IsBackend:  true,
		}
	case "React":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npm", "run", "build"},
			OutputDir:  "build",
			IsBackend:  false,
		}
	case "Next.js":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npm", "run", "build"},
			OutputDir:  ".next",
			IsBackend:  false,
		}
	case "Vue":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npm", "run", "build"},
			OutputDir:  "dist",
			IsBackend:  false,
		}
	case "Angular":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npx", "ng", "build", "--configuration=production"},
			OutputDir:  "dist",
			IsBackend:  false,
		}
	case "Svelte":
		return &models.BuildConfig{
			BaseImage:  "node:20-alpine",
			InstallCmd: []string{"npm", "ci"},
			BuildCmd:   []string{"npm", "run", "build"},
			OutputDir:  "build",
			IsBackend:  false,
		}
	default:
		return &models.BuildConfig{
			BaseImage:  "alpine:latest",
			InstallCmd: nil,
			BuildCmd:   nil,
			OutputDir:  ".",
			IsBackend:  false,
		}
	}
}

func (s *BuilderService) Build(clonedir string, projectType string, deploymentID string) (string, error, []byte, int) {
	ctx := context.Background()

	// get the build config (details)
	buildConfigDetails := GetBuildConfig(projectType)

	// create a docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		slog.Error(
			"Error while creating Docker client",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while creating Docker client", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// pull the image
	reader, err := dockerClient.ImagePull(ctx, buildConfigDetails.BaseImage, image.PullOptions{})

	if err != nil {
		slog.Error(
			"Error while pulling Docker image",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while pulling Docker image", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// drain the reader
	io.Copy(io.Discard, reader)
	reader.Close()

	// create a container
	var shellCommand string

	if buildConfigDetails.InstallCmd != nil && buildConfigDetails.BuildCmd != nil {
		shellCommand = strings.Join(buildConfigDetails.InstallCmd, " ") + " && " + strings.Join(buildConfigDetails.BuildCmd, " ")
	} else if buildConfigDetails.InstallCmd != nil {
		shellCommand = strings.Join(buildConfigDetails.InstallCmd, " ")
	} else if buildConfigDetails.BuildCmd != nil {
		shellCommand = strings.Join(buildConfigDetails.BuildCmd, " ")
	} else {
		return clonedir, nil, nil, http.StatusOK
	}

	containerResponse, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:      buildConfigDetails.BaseImage,
			Cmd:        []string{"sh", "-c", shellCommand},
			WorkingDir: "/app",
			Env:        []string{"NODE_OPTIONS=--openssl-legacy-provider"},
		},
		&container.HostConfig{
			Binds: []string{clonedir + ":/app"},
		},
		nil,
		nil,
		"relay-build-"+deploymentID,
	)

	if err != nil {
		slog.Error(
			"Error while creating the container",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while creating the container", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// enforcing a 10 min wait
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// start the container
	err = dockerClient.ContainerStart(ctx, containerResponse.ID, container.StartOptions{})

	if err != nil {
		slog.Error(
			"Error while starting the container",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while starting the container", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// stream the logs
	slog.Debug("Printing the Build & Start Logs:")
	logReader, err := dockerClient.ContainerLogs(ctx, containerResponse.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})

	if err != nil {
		slog.Error("Error while fetching container logs", slog.Any("Error", err))
		errJsonData, internalServerError := errors.NewInternalServerError("Error while fetching container logs", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}
	defer logReader.Close()

	// wait for the app to start
	statusCh, errCh := dockerClient.ContainerWait(ctx, containerResponse.ID, container.WaitConditionNotRunning)

	// stream logs
	io.Copy(os.Stdout, logReader)

	// check for exit code
	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("Error while waiting for container", slog.Any("Error", err))
			errJsonData, internalServerError := errors.NewInternalServerError("Error while waiting for container", err)
			return "", internalServerError, errJsonData, internalServerError.Code
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			slog.Error("Build failed", slog.Int64("ExitCode", status.StatusCode))
			errJsonData, internalServerError := errors.NewInternalServerError("Build failed with non-zero exit code", nil)
			return "", internalServerError, errJsonData, internalServerError.Code
		}
	}

	// clean up the container
	dockerClient.ContainerRemove(ctx, containerResponse.ID, container.RemoveOptions{Force: true})

	// return the container ID for backend projects, clone dir for frontend
	if buildConfigDetails.IsBackend {
		return containerResponse.ID, nil, nil, http.StatusOK
	} else {
		return clonedir, nil, nil, http.StatusOK
	}
}
