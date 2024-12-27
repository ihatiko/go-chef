package main

import (
	gochefcodegenutils "github.com/ihatiko/go-chef-code-gen-utils"
	"log/slog"
	"os"
	"strings"
)

const corePathPackage = "github.com/ihatiko/gochef/cli/go-chef-core"
const coreNamePackage = "go-chef-core"

func main() {
	params := strings.Join(os.Args[1:], " ")
	//TODO timeout on update
	gochefcodegenutils.AutoUpdate(corePathPackage)
	composer := gochefcodegenutils.NewExecutor()
	result, err := composer.ExecDefaultCommand(coreNamePackage)
	if err != nil {
		slog.Error("Error executing command: ", slog.Any("error", err), slog.String("command", params))
	}
	slog.Info(result.String())
}
