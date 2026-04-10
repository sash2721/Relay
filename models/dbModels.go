package models

type Users struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	Country      string `json:"country"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"passwordHash"`
	Provider     string `json:"provider"`
	CreatedAt    string `json:"createdAt"`
}

type Projects struct {
	Id                 string `json:"id"`
	UserId             string `json:"userId"`
	ProjectName        string `json:"projectName"`
	RepoURL            string `json:"repoUrl"`
	ProjectType        string `json:"projectType"`
	ActiveDeploymentId string `json:"activeDeploymentId"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type Deployments struct {
	Id            string `json:"id"`
	ProjectId     string `json:"projectId"`
	Status        string `json:"status"`
	DeployedURL   string `json:"deployedURL"`
	Subdomain     string `json:"subdomain"`
	FailureReason string `json:"failureReason"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type DeploymentLogs struct {
	Id           string `json:"id"`
	DeploymentId string `json:"deploymentId"`
	Message      string `json:"message"`
	CreatedAt    string `json:"createdAt"`
}
