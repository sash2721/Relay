package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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
