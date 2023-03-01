package main

import (
	"awesomeProject/models"
	"awesomeProject/pkg/config"
	"bytes"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/iancoleman/strcase"
	dynamic_struct "github.com/ihatiko/dynamic-struct"
	"github.com/ihatiko/log"
	"github.com/spf13/viper"
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
		log.Error("unable to decode into struct, %v", err)
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
type EnvironmentConfig struct {
	ProjectName string `env:"PROJECT_NAME"`
	ProjectPath string `env:"PROJECT_PATH"`
}

func main() {
	environmentConfig := EnvironmentConfig{}
	if err := env.Parse(&environmentConfig); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println(environmentConfig)
	return
	logCfg := log.Config{
		Caller:   false,
		DevMode:  false,
		Encoding: "console",
		Level:    "debug",
	}
	logCfg.SetConfiguration("go-chef")

	cfg, err := LoadConfig(templateYml)
	if err != nil {
		log.Fatal(err)
	}
	config, err := ParseConfig(cfg)
	if err != nil {
		panic(err)
	}

	cmp := GetExternalComponents(config)
	FillRequiredComponents(&cmp)
	env := dynamic_struct.ConstructStruct(map[string]any{
		"Grpc":        false, //TODO
		"ProjectName": config.Tree.Settings.ProjectName,
		"ServiceName": strings.ToLower(strcase.ToSnake(config.Tree.Settings.ProjectName)),
	})
	BuildTree(config.Tree, env)

	//TODO constants
	CommandsComposer(config,
		NewCommand(fmt.Sprintf(goModInit, config.Tree.Settings.ProjectName), true),
		NewCommand(goFmt, true),
		NewCommand(goModTidy, false),
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

type Command struct {
	cmd         string
	skipOnError bool
}

func NewCommand(cmd string, skipOnError bool) Command {
	return Command{cmd: cmd, skipOnError: skipOnError}
}

func CommandsComposer(config *Config, commands ...Command) {
	consoleEnv := "bash"
	if os.Getenv("GOOS") == "windows" || strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		consoleEnv = "powershell"
	}
	for _, command := range commands {
		err := ExecCommand(config, command.cmd, consoleEnv)
		log.Info(command.cmd)
		if err != nil && !command.skipOnError {
			log.Error(err)
			break
		}
	}
}
func ExecCommand(config *Config, command string, consoleEnv string) error {
	cmdFolder := exec.Command(consoleEnv, "-c", command)
	var out bytes.Buffer
	cmdFolder.Stdin = strings.NewReader("some input")
	cmdFolder.Stderr = &out
	cmdFolder.Dir = config.Tree.Settings.ProjectPath
	cmdFolder.Run()
	if cmdFolder.ProcessState.ExitCode() > 0 {
		return errors.New(out.String())
	}
	return nil
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
