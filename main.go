package main

import (
	"awesomeProject/models"
	"awesomeProject/pkg/config"
	"errors"
	"fmt"
	dynamic_struct "github.com/ihatiko/dynamic-struct"
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
	components    = "components"
	templateYml   = "template.yml"
	configuration = "config"
	mainPackage   = "main"
)

type GolangFile struct {
	Package string
	Env     interface{}
}
type Config struct {
	Tree *models.Tree
}

func LoadConfig(filename string) (*config.Config, error) {
	cfg := config.New(viper.New())
	cfg.SetConfigName(filename)
	cfg.AddConfigPath(".")
	cfg.AutomaticEnv()
	cfg.SetConfigType(yml)
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
	env := dynamic_struct.ConstructStruct(map[string]any{"ProjectName": config.Tree.Settings.ProjectName})
	BuildTree(config.Tree, env)

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

func BuildTree(tree *models.Tree, env any) {
	BuildRootFiles(tree.Settings.ProjectPath, tree.RootComponent, env)
	for _, nd := range tree.RootComponent.Nodes {
		BuildNodes(tree.Settings.ProjectPath, nd, tree.Settings, env)
		BuildFiles(tree.Settings.ProjectPath, nd, tree.Settings, env)
	}
	for _, nd := range tree.DomainComponents {
		BuildNodes(tree.Settings.ProjectPath, nd, tree.Settings, env)
	}
}

func BuildNodes(path string, node *models.Node, settings *models.Settings, env any) {
	if len(node.Nodes) > 0 || len(node.GeneratedFiles) > 0 {
		os.Mkdir(filepath.Join(path, node.Name), os.ModePerm)
	}
	for _, nd := range node.Nodes {
		BuildNodes(filepath.Join(path, node.Name), nd, settings, env)
		BuildFiles(filepath.Join(path, node.Name), nd, settings, env)
	}
}

func BuildRootFiles(path string, node *models.RootNode, obj any) {
	for _, file := range node.GeneratedFiles {
		b, err := os.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))
		if err != nil {
			panic(err)
		}
		t, err := template.New("").Parse(string(b))
		fileName := file.Name
		if file.Extension != "" {
			fileName = fmt.Sprintf("%s.%s", file.Name, file.Extension)
		}
		f, err := os.Create(filepath.Join(path, fileName))
		if err != nil {
			panic(err)
		}
		obj = dynamic_struct.ReconstructStruct(obj, dynamic_struct.Field{
			Name:  "Package",
			Value: mainPackage,
		})
		err = t.ExecuteTemplate(f, "", obj)
		if err != nil {
			panic(err)
		}
	}
}

func BuildFiles(path string, node *models.Node, settings *models.Settings, obj any) {
	for _, file := range node.GeneratedFiles {
		b, err := os.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))
		if err != nil {
			panic(err)
		}
		t, err := template.New("").Parse(string(b))
		p := filepath.Join(path, node.Name, fmt.Sprintf("%s.%s", file.Name, file.Extension))
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}
		obj = dynamic_struct.ReconstructStruct(obj, dynamic_struct.Field{
			Name:  "Package",
			Value: strings.Replace(node.Name, "-", "_", 1),
		})
		obj = FillRootSettings(file, settings, obj)
		err = t.ExecuteTemplate(f, "", obj)
		if err != nil {
			panic(err)
		}
	}
}

func FillRootSettings(file models.GeneratedFile, settings *models.Settings, obj any) any {
	if file.Name == configuration && file.Extension == yml {
		var configYamlData []string
		for _, ext := range settings.ExternalComponents {
			b, err := os.ReadFile(fmt.Sprintf("components/%s/%s", ext, "config.yml"))
			if err != nil {
				panic(err)
			}
			configYamlData = append(configYamlData, fmt.Sprintf("%s\n", string(b)))
		}
		obj = dynamic_struct.ReconstructStruct(obj, dynamic_struct.Field{
			Name:  "LogFile",
			Value: configYamlData,
		})
	}
	return obj
}
