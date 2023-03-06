package models

type Tree struct {
	DomainComponents []*Node `yaml:"domain-components" mapstructure:"domain-components"`
	Settings         *Settings
	RootComponent    *RootNode `yaml:"root-component" mapstructure:"root-component"`
}

type Settings struct {
	ExternalComponents []string `yaml:"external-components" mapstructure:"external-components"`
	ProjectSettings    *EnvironmentConfig
	DomainSettings     *DomainConfig
}
type DomainConfig struct {
}
type EnvironmentConfig struct {
	ProjectName string `validate:"required"`
	ProjectPath string `validate:"required"`
}
type Config struct {
	Tree *Tree
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
