package config

type ZarfFile struct {
	Source     string   `yaml:"source"`
	Shasum     string   `yaml:"shasum"`
	Target     string   `yaml:"target"`
	Executable bool     `yaml:"executable"`
	Symlinks   []string `yaml:"symlinks"`
}

type ZarfChart struct {
	Name      string `yaml:"name"`
	Url       string `yaml:"url"`
	Version   string `yaml:"version"`
	Namespace string `yaml:"namespace"`
	Values    string `yaml:"values"`
}

type ZarfComponent struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Default     bool        `yaml:"default"`
	Required    bool        `yaml:"required"`
	Manifests   string      `yaml:"manifests"`
	Images      []string    `yaml:"images"`
	Repos       []string    `yaml:"repos"`
	Charts      []ZarfChart `yaml:"charts"`
	Files       []ZarfFile  `yaml:"files"`
	Scripts     struct {
		PreDeploy  []string `yaml:"preDeploy"`
		PostDeploy []string `yaml:"postDeploy"`
	} `yaml:"scripts"`
}

type ZarfMetatdata struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Version      string `yaml:"version"`
	Uncompressed bool   `yaml:"uncompressed"`
}

type ZarfContainerTarget struct {
	Namespace string `yaml:"namespace"`
	Selector  string `yaml:"selector"`
	Container string `yaml:"container"`
	Path      string `yaml:"path"`
}

type ZarfData struct {
	Source string              `yaml:"source"`
	Target ZarfContainerTarget `yaml:"target"`
}

type ZarfBuildData struct {
	Terminal  string `yaml:"terminal"`
	User      string `yaml:"user"`
	Timestamp string `yaml:"timestamp"`
	Version   string `yaml:"string"`
}

type ZarfConfig struct {
	Kind       string          `yaml:"kind"`
	Metadata   ZarfMetatdata   `yaml:"metadata"`
	Package    ZarfBuildData   `yaml:"package"`
	Data       []ZarfData      `yaml:"data"`
	Components []ZarfComponent `yaml:"components"`
}
