package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sash2721/Relay/configs"
	"github.com/sash2721/Relay/errors"
)

type StorageService struct{}

func NewStorageService() *StorageService {
	return &StorageService{}
}

func Store(deploymentID string, sourceDir string, outputDir string) (string, error, []byte, int) {
	// load the server configs
	serverConfig := configs.GetServerConfig()

	// create the source & target path
	sourcePath := filepath.Join(sourceDir, outputDir)
	targetPath := filepath.Join(serverConfig.ArtifactsDir, deploymentID)

	// create the target directory
	err := os.MkdirAll(targetPath, 0755)

	if err != nil {
		slog.Error(
			"Error while creating local artifacts directory for the build",
			slog.String("DeploymentID: ", deploymentID),
			slog.Any("Error: ", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while creating local artifacts directory for the build", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	// recursively copy all the files from source path to target path
	err = copyAll(sourcePath, targetPath)

	if err != nil {
		slog.Error(
			"Error while copy files from source to destination",
			slog.String("DeploymentID: ", deploymentID),
			slog.Any("Error: ", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Error while copy files from source to destination", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	return targetPath, nil, nil, http.StatusOK
}

func Delete(deploymentID string) (error, []byte, int) {
	// load the serverconfig
	serverconfig := configs.GetServerConfig()

	// form the delete path (path to be deleted)
	deletePath := filepath.Join(serverconfig.ArtifactsDir, deploymentID)

	// delete the directory on the path
	err := os.RemoveAll(deletePath)

	if err != nil {
		slog.Error(
			"Failed to delete artifacts on the delete path",
			slog.Any("Error:", err),
		)
		errJsonData, internalServerError := errors.NewInternalServerError("Failed to delete artifacts on the delete path", err)
		return internalServerError, errJsonData, internalServerError.Code
	}

	return nil, nil, http.StatusOK
}

func GetServePath(deploymentID string) (string, error, []byte, int) {
	// load the serverconfig
	serverconfig := configs.GetServerConfig()

	// forming the serve path
	servePath := filepath.Join(serverconfig.ArtifactsDir, deploymentID)

	// verifying if the path is legit or not
	isDir, err := isDirectory(servePath)

	if err != nil {
		slog.Error(
			"Error while veryfing the path existence",
			slog.String("Path:", servePath),
			slog.Any("Error:", err),
		)

		errJsonData, internalServerError := errors.NewInternalServerError("Error while veryfing the path existence", err)
		return "", internalServerError, errJsonData, internalServerError.Code
	}

	if err != nil || !isDir {
		errJsonData, notFoundError := errors.NewNotFoundError("Artifacts not found for this deployment", nil)
		return "", notFoundError, errJsonData, notFoundError.Code
	}

	return servePath, nil, nil, http.StatusOK
}

// helper functions
func copyAll(sourcePath string, targetPath string) error {
	// get the info about the source
	sourceInfo, err := os.Stat(sourcePath)

	if err != nil {
		slog.Error(
			"Failed to stat source info",
			slog.Any("Error:", err),
		)
		return err
	}

	// if the source is a file, copy it directly
	if !sourceInfo.IsDir() {
		return copyFile(sourcePath, targetPath)
	}

	// read all the entries in the source directory
	entries, err := os.ReadDir(sourcePath)

	if err != nil {
		slog.Error(
			"Failed to read source directory",
			slog.Any("Error:", err),
		)
		return err
	}

	// recursively copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(sourcePath, entry.Name())
		destPath := filepath.Join(targetPath, entry.Name())

		if entry.IsDir() {
			os.MkdirAll(destPath, 0755)
			err := copyAll(srcPath, destPath)

			if err != nil {
				slog.Error(
					"Failed to copy the content from source to destination",
					slog.Any("Error:", err),
				)
				return err
			}
		} else {
			err := copyFile(srcPath, destPath)

			if err != nil {
				slog.Error(
					"Failed to copy the content from source to destination",
					slog.Any("Error:", err),
				)
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	// open the source file
	srcFile, err := os.Open(src)

	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// get source file info for permissions
	srcInfo, err := srcFile.Stat()

	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// create the destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())

	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// copy the contents
	_, err = io.Copy(dstFile, srcFile)

	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

func isDirectory(dirPath string) (bool, error) {
	dirInfo, err := os.Stat(dirPath)

	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("The path does not exist: %s", dirPath)
		}

		return false, fmt.Errorf("Failed to stat the path: %w", err)
	}

	return dirInfo.IsDir(), nil
}
