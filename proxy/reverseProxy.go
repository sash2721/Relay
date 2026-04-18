package proxy

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sash2721/Relay/configs"
	"github.com/sash2721/Relay/repositories"
)

type ProxyHandler struct{}

func NewProxyHandler(depRepo *repositories.DeploymentRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		subdomain := extractSubdomain(host)

		deployment, err := depRepo.GetDeploymentBySubdomain(subdomain)

		if err != nil || deployment == nil {
			http.Error(w, "Site not found", 404)
			return
		}

		if deployment.Status != "live" {
			http.Error(w, "Site is not live yet", 503)
			return
		}

		servePath := filepath.Join(configs.GetServerConfig().ArtifactsDir, deployment.Id)
		http.FileServer(http.Dir(servePath)).ServeHTTP(w, r)
	})
}

// helper function
func extractSubdomain(host string) string {
	before, _, _ := strings.Cut(host, ".")

	return before
}
