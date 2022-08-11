package main

import (
	"awesomeProject/models"
	"awesomeProject/pkg/config"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	goModTidy = "go mod tidy"
	goFmt     = "go fmt"
	goModInit = "go mod init %s"
)
const (
	yml = "yml"
)
const (
	components  = "components"
	templateYml = "template.yml"
	mainPackage = "main"
)

type GolangFile struct {
	Package    string
	Dependency interface{}
}
type Config struct {
	Tree *models.Tree
}

func LoadConfig(filename string) (*config.Config, error) {
	cfg := config.New(viper.New())
	cfg.SetConfigName(filename)
	cfg.AddConfigPath(".")
	cfg.AutomaticEnv()
	cfg.SetConfigType("yml")
	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}

		return nil, err
	}

	return cfg, nil
}
func ParseConfig(v *config.Config) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}

type ExternalComponent struct {
	Config          string
	Constructor     string
	ObjectName      string
	LogErrorContent string
	External        bool
}

func main() {
	cfg, err := LoadConfig(templateYml)
	if err != nil {
		panic(err)
	}
	config, err := ParseConfig(cfg)
	if err != nil {
		panic(err)
	}
	cmp := GetExternalComponents(config)
	FillRequiredComponents(&cmp)

	BuildTree(config.Tree)

	//TODO constants
	CommandsComposer(config,
		goFmt,
		fmt.Sprintf(goModInit, config.Tree.Settings.ProjectName),
		goModTidy,
	)
}
func FillRequiredComponents(components *map[string]ExternalComponent) {
	fmt.Println()
}
func GetExternalComponents(config *Config) map[string]ExternalComponent {
	cpm := make(map[string]ExternalComponent)
	for _, externalComponent := range config.Tree.Settings.ExternalComponents {
		files, err := os.ReadDir(filepath.Join(components, externalComponent))
		if err != nil {
			panic(err)
		}
		if len(files) == 0 {
			continue
		}
		extComponent := ExternalComponent{}
		for _, file := range files {
			f, err := os.ReadFile(filepath.Join(components, externalComponent, file.Name()))
			if err != nil {
				panic(err)
			}

			extension := strings.Split(file.Name(), ".")[1]
			if extension == yml {
				extComponent.Config = string(f)
				continue
			}
		}
		extComponent.Constructor =
			fmt.Sprintf("%s,err := config.%s.NewComponent()",
				externalComponent,
				strings.ToTitle(externalComponent),
			)
		extComponent.LogErrorContent = GetComponentLogError()
		cpm[externalComponent] = extComponent
	}

	return cpm
}

func GetComponentLogError() string {
	errMsg := `
	if err != nil {
		log.Fatal(err)
	}
`
	return errMsg
}

func CommandsComposer(config *Config, commands ...string) {
	for _, command := range commands {
		ExecCommand(config, command)
	}
}
func ExecCommand(config *Config, command string) {
	cmdFolder := exec.Command("bash", "-c", command)
	cmdFolder.Dir = config.Tree.Settings.ProjectPath
	cmdFolder.Run()
}

func BuildTree(tree *models.Tree) {
	BuildRootFiles(tree.Settings.ProjectPath, tree.RootComponent, tree.Settings)
	for _, nd := range tree.RootComponent.Nodes {
		BuildNodes(tree.Settings.ProjectPath, nd, tree.Settings)
		BuildFiles(tree.Settings.ProjectPath, nd, tree.Settings)
	}
	for _, nd := range tree.DomainComponents {
		BuildNodes(tree.Settings.ProjectPath, nd, tree.Settings)
	}
}

func BuildNodes(path string, node *models.Node, settings *models.Settings) {
	if len(node.Nodes) > 0 || len(node.GeneratedFiles) > 0 {
		os.Mkdir(filepath.Join(path, node.Name), os.ModePerm)
	}
	for _, nd := range node.Nodes {
		BuildNodes(filepath.Join(path, node.Name), nd, settings)
		BuildFiles(filepath.Join(path, node.Name), nd, settings)
	}
}

func BuildRootFiles(path string, node *models.RootNode, settings *models.Settings) {
	for _, file := range node.GeneratedFiles {
		b, err := os.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))
		if err != nil {
			fmt.Errorf(err.Error())
		}
		t, err := template.New("").Parse(string(b))
		fileName := file.Name
		if file.Extension != "" {
			fileName = fmt.Sprintf("%s.%s", file.Name, file.Extension)
		}
		f, err := os.Create(filepath.Join(path, fileName))
		if err != nil {
			fmt.Errorf(err.Error())
		}
		golangFile := GolangFile{
			Package:    mainPackage,
			Dependency: settings,
		}
		err = t.ExecuteTemplate(f, "", golangFile)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
}

func BuildFiles(path string, node *models.Node, settings *models.Settings) {
	for _, file := range node.GeneratedFiles {
		b, err := os.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))
		if err != nil {
			fmt.Errorf(err.Error())
		}
		t, err := template.New("").Parse(string(b))
		f, err := os.Create(filepath.Join(path, node.Name, fmt.Sprintf("%s.%s", file.Name, file.Extension)))
		if err != nil {
			fmt.Errorf(err.Error())
		}
		golangFile := GolangFile{
			Package: strings.Replace(node.Name, "-", "_", 1),
		}
		err = t.ExecuteTemplate(f, "", golangFile)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
}
