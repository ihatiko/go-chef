package main

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ihatiko/go-chef/commands/project"
	"github.com/ihatiko/go-chef/models"
	"github.com/ihatiko/go-chef/ui"
	"github.com/ihatiko/log"
	"os"
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

func FillFlags(args []string, cfg *models.EnvironmentConfig) *models.EnvironmentConfig {
	//TODO сделать нормально
	for _, arg := range args {
		formattedArg := strings.Split(arg, "=")
		if len(formattedArg) < 1 {
			continue
		}
		switch strings.ToUpper(formattedArg[0]) {
		case "--PROJECT_PATH":
			cfg.ProjectPath = formattedArg[1]
		case "--PROJECT_NAME":
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
	CommandProcess(os.Args)
}
