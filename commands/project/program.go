package project

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	dynamic_struct "github.com/ihatiko/dynamic-struct"
	"github.com/ihatiko/go-chef/models"
	config_parser "github.com/ihatiko/go-chef/parse/config-parser"
	"github.com/ihatiko/log"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	yml = "yml"
)

const (
	goModTidy = "go mod tidy"
	goFmt     = "go fmt"
	goModInit = "go mod init %s"
)

const (
	components    = "components"
	templateYml   = "template.yml"
	configuration = "config"
	mainPackage   = "main"
)

//go:embed template.yml
var fTemplate []byte

//go:embed templates
var templates embed.FS

func NewCommand(cmd string, skipOnError bool) Command {
	return Command{cmd: cmd, skipOnError: skipOnError}
}

type Command struct {
	cmd         string
	skipOnError bool
}

func Mkdir(envConfig *models.EnvironmentConfig) error {
	var err error
	_, err = os.ReadDir(envConfig.ProjectPath)
	if errors.Is(err, fs.ErrNotExist) {
		err := os.Mkdir(envConfig.ProjectPath, os.ModePerm)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	return err
}

func CommandsComposer(config *models.Config, commands ...Command) {
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

func ExecCommand(config *models.Config, command string, consoleEnv string) error {
	cmdFolder := exec.Command(consoleEnv, "-c", command)
	var out bytes.Buffer
	cmdFolder.Stdin = strings.NewReader("some input")
	cmdFolder.Stderr = &out
	cmdFolder.Dir = config.Tree.Settings.ProjectSettings.ProjectPath
	cmdFolder.Run()
	if cmdFolder.ProcessState.ExitCode() > 0 {
		return errors.New(out.String())
	}
	return nil
}

func BuildTree(tree *models.Tree, env any) {
	BuildRootFiles(tree.Settings.ProjectSettings.ProjectPath, tree.RootComponent, env)
	for _, nd := range tree.RootComponent.Nodes {
		BuildNodes(tree.Settings.ProjectSettings.ProjectPath, nd, env)
		BuildFiles(tree.Settings.ProjectSettings.ProjectPath, nd, env)
	}
	for _, nd := range tree.DomainComponents {
		BuildNodes(tree.Settings.ProjectSettings.ProjectPath, nd, env)
	}
}

func BuildNodes(path string, node *models.Node, env any) {
	if len(node.Nodes) > 0 || len(node.GeneratedFiles) > 0 {
		os.Mkdir(filepath.Join(path, node.Name), os.ModePerm)
	}
	for _, nd := range node.Nodes {
		BuildNodes(filepath.Join(path, node.Name), nd, env)
		BuildFiles(filepath.Join(path, node.Name), nd, env)
	}
}

func BuildRootFiles(path string, node *models.RootNode, obj any) {
	for _, file := range node.GeneratedFiles {
		b, err := templates.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))
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

func BuildFiles(path string, node *models.Node, obj any) {
	for _, file := range node.GeneratedFiles {
		b, err := templates.ReadFile(fmt.Sprintf("templates/%s.tmpl", file.Template))

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
		obj = FillRootSettings(file, obj)
		err = t.ExecuteTemplate(f, "", obj)
		if err != nil {
			panic(err)
		}
	}
}

func FillRootSettings(file models.GeneratedFile, obj any) any {
	if file.Name == configuration && file.Extension == yml {
		var configYamlData []string
		/*		for _, ext := range settings.ExternalComponents {
				//TODO перейти на систему динамических модулей
				if ext == "log" {
					b, err := os.ReadFile(fmt.Sprintf("components/%s/%s", ext, "config.yml"))
					if err != nil {
						panic(err)
					}
					configYamlData = append(configYamlData, fmt.Sprintf("%s\n", string(b)))
				}
			}*/
		obj = dynamic_struct.ReconstructStruct(obj, dynamic_struct.Field{
			Name:  "LogFile",
			Value: configYamlData,
		})
	}
	return obj
}

func BuildProjectProgram(envConfig *models.EnvironmentConfig) {
	cfg, err := config_parser.LoadConfig(fTemplate)
	if err != nil {
		log.Fatal(err)
	}
	config, err := config_parser.ParseConfig[models.Config](cfg)
	if err != nil {
		log.Fatal(err)
	}
	config.Tree.Settings.ProjectSettings = envConfig
	err = Mkdir(envConfig)
	if err != nil {
		log.Error(err)
		return
	}
	env := dynamic_struct.ConstructStruct(map[string]any{
		"Grpc":        false, //TODO
		"ProjectName": envConfig.ProjectName,
		"ServiceName": strings.ToLower(strcase.ToSnake(envConfig.ProjectName)),
	})
	BuildTree(config.Tree, env)

	CommandsComposer(config,
		NewCommand(fmt.Sprintf(goModInit, config.Tree.Settings.ProjectSettings.ProjectName), true),
		NewCommand(goFmt, true),
		NewCommand(goModTidy, false),
	)
}
