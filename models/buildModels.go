package models

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

type BuildConfig struct {
	BaseImage  string
	InstallCmd []string
	BuildCmd   []string
	OutputDir  string
	IsBackend  bool
}
