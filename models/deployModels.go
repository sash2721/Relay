package models

type DeploymentResponse struct {
	Id            string `json:"id"`
	ProjectId     string `json:"projectId"`
	Status        string `json:"status"`
	DeployedURL   string `json:"deployedUrl"`
	Subdomain     string `json:"subdomain"`
	FailureReason string `json:"failureReason"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type DeploymentListResponse struct {
	Deployments []DeploymentResponse `json:"deployments"`
	Count       int                  `json:"count"`
}
