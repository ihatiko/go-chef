package main

import (
	"bufio"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ihatiko/go-chef/commands/project"
	"github.com/ihatiko/go-chef/models"
	"github.com/ihatiko/go-chef/ui"
	"github.com/ihatiko/log"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidCommand = errors.New("invalid command. Please use --help to get more info")

const HelpTemplate = `
	- go-chef
		interactive ui

	- go-chef cook-project
		create project by clean architecture
			--PROJECT_PATH 
				a path to the project
			--PROJECT_NAME
				project name

	- go-chef --help
		get a project description
`
const (
	PROJECT_PATH = "PROJECT_PATH"
	PROJECT_NAME = "PROJECT_NAME"
)

func FillEnvironment(args []string, cfg *models.EnvironmentConfig) *models.EnvironmentConfig {
	//TODO сделать нормально
	for _, arg := range args {
		arg = strings.Replace(arg, "--", "", 1)
		formattedArg := strings.Split(arg, "=")
		if len(formattedArg) < 1 {
			continue
		}
		switch strings.ToUpper(formattedArg[0]) {
		case PROJECT_PATH:
			cfg.ProjectPath = formattedArg[1]
		case PROJECT_NAME:
			cfg.ProjectName = formattedArg[1]
		}

	}
	return cfg
}

func FillFlags(args []string, cfg *models.EnvironmentConfig) *models.EnvironmentConfig {
	//TODO сделать нормально
	for _, arg := range args {
		arg = strings.Replace(arg, "--", "", 1)
		formattedArg := strings.Split(arg, "=")
		if len(formattedArg) < 1 {
			continue
		}
		switch strings.ToUpper(formattedArg[0]) {
		case "PROJECT_PATH":
			cfg.ProjectPath = formattedArg[1]
		case "PROJECT_NAME":
			cfg.ProjectName = formattedArg[1]
		}

	}
	return cfg
}

func CommandProcess(args []string) {
	if len(args) == 1 {
		ui.StartUI(CreateProject)
		return
	}
	if len(args) < 2 {
		fmt.Printf(HelpTemplate)
	}
	//TODO consts and move to new folder
	switch strings.ToLower(args[1]) {
	case "cook-project":
		CreateProject(args[1:])
	case "--help":
		fmt.Printf(HelpTemplate)
	default:
		fmt.Println(ErrInvalidCommand)
	}
	return
}
func CreateProject(args []string) {
	createProjectConfig := &models.EnvironmentConfig{}
	createProjectConfig = FillFlags(args, createProjectConfig)
	validate := validator.New()
	err := validate.Struct(createProjectConfig)
	if err != nil {
		log.Error(err)
		return
	}
	project.BuildProjectProgram(createProjectConfig)
}
func main() {
	logCfg := log.Config{
		Caller:   false,
		DevMode:  false,
		Encoding: "console",
		Level:    "debug",
	}
	logCfg.SetConfiguration("go-chef")
	//CommandProcess(os.Args)
	path := `C:\testProject`
	domainName := "test-domain"
	packageName := strings.ReplaceAll(domainName, "-", "_")
	projectName, err := gerProjectName(path)
	if err != nil {
		log.Fatal(err)
	}
	dirs, err := os.ReadDir(filepath.Join(path, "internal"))
	if err != nil {
		log.Fatal(err)
	}
	if !checkDir(dirs, "server") {
		log.Fatal("folder server does not exist")
	}
	if err := upsertFeaturesFolder(dirs, path); err != nil {
		log.Fatal(err)
	}
	dirs, err = os.ReadDir(filepath.Join(path, "internal/features"))
	if err != nil {
		log.Fatal(err)
	}
	if findDomain(dirs, domainName) {
		return
	}
	err = os.Mkdir(filepath.Join(path, "internal/features", domainName), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(packageName, projectName)
}

func findDomain(dirs []os.DirEntry, data string) bool {
	for _, dir := range dirs {
		if dir.Name() == data {
			return true
		}
	}

	return false
}

func upsertFeaturesFolder(dirs []os.DirEntry, path string) error {
	if !checkDir(dirs, "features") {
		return os.Mkdir(filepath.Join(path, "internal/features"), os.ModePerm)
	}
	return nil
}
func checkDir(dirs []os.DirEntry, name string) bool {
	for _, dir := range dirs {
		if dir.Name() == name {
			return true
		}
	}
	return false
}

//go:embed domain
var templates embed.FS

func gerProjectName(path string) (string, error) {
	combinedPath := filepath.Join(path, "go.mod")
	f, err := os.Open(combinedPath)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(f)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	projectName := strings.Replace(string(line), "module ", "", 1)
	return projectName, nil
}
