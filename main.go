package main

import (
	gochefcodegenutils "github.com/ihatiko/go-chef-code-gen-utils"
	"log/slog"
	"os"
	"strings"
)

const corePathPackage = "github.com/ihatiko/go-chef-proxy"
const coreNamePackage = "go-chef-proxy"

func main() {
	params := strings.Join(os.Args[1:], " ")
	//TODO timeout on update and tmp call
	gochefcodegenutils.AutoUpdate(corePathPackage)
	composer := gochefcodegenutils.NewExecutor()
	result, err := composer.ExecDefaultCommand(coreNamePackage)
	if err != nil {
		slog.Error("Error executing command: ", slog.Any("error", err), slog.String("command", params))
	}
	slog.Info(result.String())
}
