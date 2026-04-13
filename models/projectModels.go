package models

type CreateProjectRequest struct {
	ProjectName string `json:"projectName"`
	RepoURL     string `json:"repoUrl"`
}

type ProjectResponse struct {
	Id                 string `json:"id"`
	ProjectName        string `json:"projectName"`
	RepoURL            string `json:"repoUrl"`
	ProjectType        string `json:"projectType"`
	ActiveDeploymentId string `json:"activeDeploymentId"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Count    int               `json:"count"`
}
