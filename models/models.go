package models

type Tree struct {
	DomainComponents []*Node `yaml:"domain-components" mapstructure:"domain-components"`
	Settings         *Settings
	RootComponent    *RootNode `yaml:"root-component" mapstructure:"root-component"`
}

type Settings struct {
	ExternalComponents []string `yaml:"external-components" mapstructure:"external-components"`
	ProjectPath        string   `yaml:"project-path" mapstructure:"project-path"`
	ProjectName        string   `yaml:"project-name" mapstructure:"project-name"`
}

type RootNode struct {
	GeneratedFiles []*GeneratedFile `yaml:"generated-files" mapstructure:"generated-files"`
	Nodes          []*Node          `yaml:"domain-components" mapstructure:"nodes"`
}

type GeneratedFile struct {
	Name      string
	Extension Extension
	Template  string
}

type Node struct {
	Name           string
	Nodes          []*Node
	GeneratedFiles []GeneratedFile `yaml:"generated-files" mapstructure:"generated-files"`
}
